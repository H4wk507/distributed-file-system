package client

import (
	"crypto/sha256"
	"dfs-backend/dfs/common"
	"dfs-backend/dfs/streaming"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MasterClient struct {
	// Known cluster nodes (seed nodes for discovery)
	seedNodes []NodeAddr
	// Current master address (dynamically discovered)
	masterIP   string
	masterPort int
	masterMu   sync.RWMutex

	timeout  time.Duration
	clientID uuid.UUID
}

type NodeAddr struct {
	IP   string
	Port int
}

func NewMasterClient(ip string, port int) *MasterClient {
	return &MasterClient{
		seedNodes:  []NodeAddr{{IP: ip, Port: port}},
		masterIP:   ip,
		masterPort: port,
		timeout:    30 * time.Second,
		clientID:   uuid.New(),
	}
}

// NewMasterClientWithSeeds creates a client with multiple seed nodes for failover
func NewMasterClientWithSeeds(seeds []NodeAddr) *MasterClient {
	if len(seeds) == 0 {
		panic("at least one seed node is required")
	}
	return &MasterClient{
		seedNodes:  seeds,
		masterIP:   seeds[0].IP,
		masterPort: seeds[0].Port,
		timeout:    30 * time.Second,
		clientID:   uuid.New(),
	}
}

// getMasterAddr returns current master address (thread-safe)
func (c *MasterClient) getMasterAddr() (string, int) {
	c.masterMu.RLock()
	defer c.masterMu.RUnlock()
	return c.masterIP, c.masterPort
}

// setMasterAddr updates current master address (thread-safe)
func (c *MasterClient) setMasterAddr(ip string, port int) {
	c.masterMu.Lock()
	defer c.masterMu.Unlock()
	if c.masterIP != ip || c.masterPort != port {
		log.Printf("Master changed: %s:%d -> %s:%d", c.masterIP, c.masterPort, ip, port)
		c.masterIP = ip
		c.masterPort = port
	}
}

// discoverMaster queries seed nodes to find the current master
func (c *MasterClient) discoverMaster() error {
	for _, seed := range c.seedNodes {
		addr := fmt.Sprintf("%s:%d", seed.IP, seed.Port)
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			continue
		}

		// Send discovery request
		msg := common.MessageWithTime{
			Message: common.Message{
				Type: common.MessageDiscovery,
				From: c.clientID,
			},
			LogicalTime: 0,
		}

		conn.SetDeadline(time.Now().Add(5 * time.Second))
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msg); err != nil {
			conn.Close()
			continue
		}

		decoder := json.NewDecoder(conn)
		var response common.MessageWithTime
		if err := decoder.Decode(&response); err != nil {
			conn.Close()
			continue
		}
		conn.Close()

		// Parse peer list and find master
		var peers []*common.NodeInfo
		if err := json.Unmarshal(response.Message.Payload, &peers); err != nil {
			continue
		}

		for _, peer := range peers {
			if peer.Role == common.RoleMaster && peer.Status == common.StatusOnline {
				c.setMasterAddr(peer.IP, peer.Port)
				log.Printf("Discovered master: %s:%d", peer.IP, peer.Port)
				return nil
			}
		}
	}
	return fmt.Errorf("no master found in cluster")
}

// Ask the current master for the peer list and filter storage nodes
func (c *MasterClient) GetStorageNodes() ([]*common.NodeInfo, error) {
	msg := common.Message{
		Type: common.MessageDiscovery,
		From: c.clientID,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send discovery request: %w", err)
	}

	var peers []*common.NodeInfo
	if err := json.Unmarshal(response.Payload, &peers); err != nil {
		return nil, fmt.Errorf("failed to parse peers: %w", err)
	}

	var storage []*common.NodeInfo
	for _, p := range peers {
		if p.Role == common.RoleStorage && p.Status == common.StatusOnline {
			storage = append(storage, p)
		}
	}

	return storage, nil
}

func (c *MasterClient) DeleteFile(fileID uuid.UUID, hash string) (*common.APIFileDeleteResponse, error) {
	requestID := uuid.New().String()

	request := common.APIFileDeleteRequest{
		RequestID: requestID,
		FileID:    fileID,
		Hash:      hash,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal delete request: %w", err)
	}

	msg := common.Message{
		Type:    common.MessageAPIFileDelete,
		From:    c.clientID,
		Payload: payload,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send delete request: %w", err)
	}

	var deleteResponse common.APIFileDeleteResponse
	if err := json.Unmarshal(response.Payload, &deleteResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal delete response: %w", err)
	}

	return &deleteResponse, nil
}

// StreamUploadSession holds open connections for chunked streaming
type StreamUploadSession struct {
	SessionID    string
	FileID       uuid.UUID
	Size         int64
	Connections  []net.Conn
	NodeIDs      []uuid.UUID
	mu           sync.Mutex
	bytesWritten int64
}

type UploadResult struct {
	Hash    string
	NodeIDs []uuid.UUID
}

// InitStreamUpload initializes a streaming upload session and opens connections
func (c *MasterClient) InitStreamUpload(fileID uuid.UUID, filename, contentType string, size int64) (*StreamUploadSession, error) {
	request := common.StreamUploadInitRequest{
		RequestID:   uuid.New().String(),
		FileID:      fileID,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream upload request: %w", err)
	}

	msg := common.Message{
		Type:    common.MessageStreamUploadInit,
		From:    c.clientID,
		Payload: payload,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send stream upload init: %w", err)
	}

	var uploadReady common.StreamUploadReadyResponse
	if err := json.Unmarshal(response.Payload, &uploadReady); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload ready response: %w", err)
	}

	if !uploadReady.Success {
		return nil, fmt.Errorf("upload init failed: %s", uploadReady.Error)
	}

	if len(uploadReady.StorageNodes) == 0 {
		return nil, fmt.Errorf("no storage nodes available")
	}

	// Open connections to all storage nodes
	var connections []net.Conn
	var nodeIDs []uuid.UUID
	for _, node := range uploadReady.StorageNodes {
		conn, err := net.DialTimeout("tcp", node.Addr, c.timeout)
		if err != nil {
			// Close already opened connections
			for _, c := range connections {
				c.Close()
			}
			return nil, fmt.Errorf("failed to connect to storage %s: %w", node.Addr, err)
		}

		// Set long deadline for large files
		deadline := 30*time.Minute + time.Duration(size/(1024*1024*1024))*time.Minute
		conn.SetDeadline(time.Now().Add(deadline))

		// Send header
		header := &streaming.StreamHeader{
			Magic:     streaming.MagicUpload,
			SessionID: uploadReady.SessionID,
			FileSize:  size,
			ChunkSize: streaming.DefaultChunkSize,
		}

		if err := header.Encode(conn); err != nil {
			conn.Close()
			for _, c := range connections {
				c.Close()
			}
			return nil, fmt.Errorf("failed to send header to %s: %w", node.Addr, err)
		}

		connections = append(connections, conn)
		nodeIDs = append(nodeIDs, node.NodeID)
	}

	return &StreamUploadSession{
		SessionID:   uploadReady.SessionID,
		FileID:      fileID,
		Size:        size,
		Connections: connections,
		NodeIDs:     nodeIDs,
	}, nil
}

// WriteChunk writes data to all storage node connections
func (s *StreamUploadSession) WriteChunk(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, conn := range s.Connections {
		n, err := conn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write to connection %d: %w", i, err)
		}
		if n != len(data) {
			return fmt.Errorf("incomplete write to connection %d: %d/%d", i, n, len(data))
		}
	}

	s.bytesWritten += int64(len(data))
	return nil
}

// Finalize waits for responses from all storage nodes and returns the result
func (s *StreamUploadSession) Finalize() (*UploadResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bytesWritten != s.Size {
		return nil, fmt.Errorf("incomplete upload: wrote %d/%d bytes", s.bytesWritten, s.Size)
	}

	var hash string
	var firstError error
	var successfulNodes []uuid.UUID

	for i, conn := range s.Connections {
		// Read response (321 bytes: 1 status + 64 hash + 256 error)
		resp := make([]byte, 321)
		if _, err := io.ReadFull(conn, resp); err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("failed to read response from connection %d: %w", i, err)
			}
			continue
		}

		if resp[0] != 1 {
			errMsg := string(trimNull(resp[65:321]))
			if firstError == nil {
				firstError = fmt.Errorf("storage error from connection %d: %s", i, errMsg)
			}
			continue
		}

		if hash == "" {
			hash = string(trimNull(resp[1:65]))
		}
		successfulNodes = append(successfulNodes, s.NodeIDs[i])
	}

	if firstError != nil {
		return nil, firstError
	}

	return &UploadResult{
		Hash:    hash,
		NodeIDs: successfulNodes,
	}, nil
}

// Close closes all connections
func (s *StreamUploadSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, conn := range s.Connections {
		conn.Close()
	}
	s.Connections = nil
}

// UploadFileStream uploads a file using streaming (legacy single-call method)
func (c *MasterClient) UploadFileStream(fileID uuid.UUID, filename, contentType string, size int64, reader io.Reader) (*UploadResult, error) {
	session, err := c.InitStreamUpload(fileID, filename, contentType, size)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// Stream data in chunks
	buf := make([]byte, 4*1024*1024) // 4MB buffer
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if writeErr := session.WriteChunk(buf[:n]); writeErr != nil {
				return nil, writeErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}
	}

	return session.Finalize()
}

func (c *MasterClient) DownloadFileStream(fileID uuid.UUID, hash string, writer io.Writer) error {
	request := common.StreamDownloadInitRequest{
		RequestID: uuid.New().String(),
		FileID:    fileID,
		Hash:      hash,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal stream download request: %w", err)
	}

	msg := common.Message{
		Type:    common.MessageStreamDownloadInit,
		From:    c.clientID,
		Payload: payload,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return fmt.Errorf("failed to send stream download init: %w", err)
	}

	var downloadReady common.StreamDownloadReadyResponse
	if err := json.Unmarshal(response.Payload, &downloadReady); err != nil {
		return fmt.Errorf("failed to unmarshal download ready response: %w", err)
	}

	if !downloadReady.Success {
		return fmt.Errorf("download init failed: %s", downloadReady.Error)
	}

	return c.streamFromStorage(downloadReady.StorageAddr, hash, downloadReady.Size, writer)
}

func (c *MasterClient) streamFromStorage(addr, hash string, expectedSize int64, writer io.Writer) error {
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}
	defer conn.Close()

	req := &streaming.DownloadRequest{Hash: hash}
	if err := req.Encode(conn); err != nil {
		return fmt.Errorf("failed to send download request: %w", err)
	}

	header, err := streaming.DecodeHeader(conn)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	if header.FileSize < 0 {
		return c.readStorageError(conn)
	}

	if expectedSize > 0 && header.FileSize != expectedSize {
		return fmt.Errorf("size mismatch: expected %d bytes, got %d", expectedSize, header.FileSize)
	}

	copier := &streamCopier{
		conn:         conn,
		writer:       writer,
		totalSize:    header.FileSize,
		expectedHash: hash,
	}

	return copier.copy()
}

// ProgressFunc is called during download to report progress
// written: bytes downloaded so far, total: total file size
// Return error to cancel the download
type ProgressFunc func(written, total int64) error

type streamCopier struct {
	conn         net.Conn
	writer       io.Writer
	totalSize    int64
	expectedHash string

	written int64
	hasher  hash.Hash
}

func (sc *streamCopier) copy() error {
	sc.hasher = sha256.New()

	// Get buffer from pool
	bufPtr := streaming.GetBuffer()
	defer streaming.PutBuffer(bufPtr)
	buf := *bufPtr

	// Read and write loop
	for sc.written < sc.totalSize {
		// Set per-chunk deadline
		if err := sc.conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		// Calculate how much to read
		toRead := min(int64(len(buf)), sc.totalSize-sc.written)

		// Read from connection
		n, readErr := sc.conn.Read(buf[:toRead])

		// Process any data we got before handling errors
		if n > 0 {
			if err := sc.processChunk(buf[:n]); err != nil {
				return err
			}
		}

		// Handle read error
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("failed to read data at offset %d: %w", sc.written, readErr)
		}
	}

	// Verify completeness
	if sc.written != sc.totalSize {
		return fmt.Errorf("incomplete download: expected %d bytes, got %d", sc.totalSize, sc.written)
	}

	// Verify hash
	if err := sc.verifyHash(); err != nil {
		return err
	}

	return nil
}

func (sc *streamCopier) processChunk(data []byte) error {
	// Write to destination
	nw, err := sc.writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data at offset %d: %w", sc.written, err)
	}
	if nw != len(data) {
		return fmt.Errorf("short write at offset %d: wrote %d of %d bytes", sc.written, nw, len(data))
	}

	// Update hash
	if sc.hasher != nil {
		sc.hasher.Write(data)
	}

	sc.written += int64(len(data))

	return nil
}

func (sc *streamCopier) verifyHash() error {
	computedHash := hex.EncodeToString(sc.hasher.Sum(nil))

	expected := strings.ToLower(sc.expectedHash)
	computed := strings.ToLower(computedHash)

	if computed != expected {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expected, computed)
	}

	return nil
}

func (c *MasterClient) readStorageError(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	errBuf := make([]byte, 1024)
	n, _ := io.ReadAtLeast(conn, errBuf, 1)
	if n == 0 {
		return fmt.Errorf("storage error: unknown error (no message received)")
	}

	errMsg := strings.TrimRight(string(errBuf[:n]), "\x00 \t\n\r")
	return fmt.Errorf("storage error: %s", errMsg)
}

func trimNull(b []byte) []byte {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			return b[:i+1]
		}
	}
	return b[:0]
}

func (c *MasterClient) sendRequest(msg common.Message) (*common.Message, error) {
	return c.sendRequestWithRetry(msg, 2)
}

func (c *MasterClient) sendRequestWithRetry(msg common.Message, retries int) (*common.Message, error) {
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		masterIP, masterPort := c.getMasterAddr()
		addr := fmt.Sprintf("%s:%d", masterIP, masterPort)

		conn, err := net.DialTimeout("tcp", addr, c.timeout)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to master at %s: %w", addr, err)
			// Try to discover new master
			if discoverErr := c.discoverMaster(); discoverErr != nil {
				log.Printf("Master discovery failed: %v", discoverErr)
			}
			continue
		}

		conn.SetDeadline(time.Now().Add(c.timeout))

		msgWithTime := common.MessageWithTime{
			Message:     msg,
			LogicalTime: 0, // API client doesn't participate in logical clock
		}

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(msgWithTime); err != nil {
			conn.Close()
			lastErr = fmt.Errorf("failed to send message: %w", err)
			if discoverErr := c.discoverMaster(); discoverErr != nil {
				log.Printf("Master discovery failed: %v", discoverErr)
			}
			continue
		}

		decoder := json.NewDecoder(conn)
		var responseWithTime common.MessageWithTime
		if err := decoder.Decode(&responseWithTime); err != nil {
			conn.Close()
			lastErr = fmt.Errorf("failed to receive response: %w", err)
			if discoverErr := c.discoverMaster(); discoverErr != nil {
				log.Printf("Master discovery failed: %v", discoverErr)
			}
			continue
		}

		conn.Close()
		return &responseWithTime.Message, nil
	}

	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}

func (c *MasterClient) Ping() error {
	masterIP, masterPort := c.getMasterAddr()
	addr := fmt.Sprintf("%s:%d", masterIP, masterPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		// Try discovery and ping new master
		if discoverErr := c.discoverMaster(); discoverErr != nil {
			return fmt.Errorf("master not reachable and discovery failed: %w", err)
		}
		masterIP, masterPort = c.getMasterAddr()
		addr = fmt.Sprintf("%s:%d", masterIP, masterPort)
		conn, err = net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			return fmt.Errorf("new master not reachable at %s: %w", addr, err)
		}
	}
	conn.Close()
	return nil
}
