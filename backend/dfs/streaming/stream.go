package streaming

import (
	"context"
	"dfs-backend/dfs/storage"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// UploadCompleteCallback is called when an upload is successfully completed
type UploadCompleteCallback func(sessionID string, fileID string, hash string)

// StreamServer handles binary file streaming on storage nodes
type StreamServer struct {
	storage          *storage.LocalStorage
	sessions         *SessionManager
	listener         net.Listener
	port             int
	logger           *log.Logger
	stopChan         chan struct{}
	wg               sync.WaitGroup
	onUploadComplete UploadCompleteCallback
}

// NewStreamServer creates a new streaming server
func NewStreamServer(port int, localStorage *storage.LocalStorage, logger *log.Logger) *StreamServer {
	return &StreamServer{
		storage:  localStorage,
		sessions: NewSessionManager(),
		port:     port,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// GetSessionManager returns the session manager
func (s *StreamServer) GetSessionManager() *SessionManager {
	return s.sessions
}

// SetUploadCompleteCallback sets the callback for completed uploads
func (s *StreamServer) SetUploadCompleteCallback(cb UploadCompleteCallback) {
	s.onUploadComplete = cb
}

// Start starts the streaming server
func (s *StreamServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start stream server: %w", err)
	}
	s.listener = listener
	s.logger.Printf("Stream server started on port %d", s.port)

	s.sessions.StartCleanupRoutine(s.stopChan)

	go s.acceptConnections(ctx)

	return nil
}

// Stop stops the streaming server
func (s *StreamServer) Stop() error {
	close(s.stopChan)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	return nil
}

// Port returns the streaming port
func (s *StreamServer) Port() int {
	return s.port
}

func (s *StreamServer) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					s.logger.Printf("Failed to accept stream connection: %v", err)
					continue
				}
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				s.handleConnection(conn)
			}()
		}
	}
}

func (s *StreamServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read first byte to determine request type
	magicBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, magicBuf); err != nil {
		s.logger.Printf("Failed to read magic byte: %v", err)
		return
	}

	switch magicBuf[0] {
	case MagicUpload:
		s.handleUpload(conn, magicBuf[0])
	case MagicDownload:
		s.handleDownload(conn)
	default:
		s.logger.Printf("Unknown magic byte: %x", magicBuf[0])
	}
}

func (s *StreamServer) handleUpload(conn net.Conn, firstByte byte) {
	// Read rest of header (we already read the magic byte)
	restBuf := make([]byte, HeaderSize-1)
	if _, err := io.ReadFull(conn, restBuf); err != nil {
		s.logger.Printf("Failed to read header: %v", err)
		return
	}

	// Reconstruct full header buffer
	fullBuf := make([]byte, HeaderSize)
	fullBuf[0] = firstByte
	copy(fullBuf[1:], restBuf)

	// Parse header manually since we already read the bytes
	header := &StreamHeader{
		Magic:     fullBuf[0],
		SessionID: string(trimNull(fullBuf[1:37])),
	}
	header.FileSize = int64(uint64(fullBuf[37])<<56 | uint64(fullBuf[38])<<48 | uint64(fullBuf[39])<<40 | uint64(fullBuf[40])<<32 |
		uint64(fullBuf[41])<<24 | uint64(fullBuf[42])<<16 | uint64(fullBuf[43])<<8 | uint64(fullBuf[44]))
	header.ChunkSize = int32(uint32(fullBuf[45])<<24 | uint32(fullBuf[46])<<16 | uint32(fullBuf[47])<<8 | uint32(fullBuf[48]))

	// Validate session with retry (session might be propagating from master)
	var session *UploadSession
	var exists bool
	for retry := 0; retry < 10; retry++ {
		session, exists = s.sessions.GetSession(header.SessionID)
		if exists {
			break
		}
		if retry < 9 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	if !exists {
		s.logger.Printf("Invalid or expired session: %s", header.SessionID)
		s.sendUploadResponse(conn, false, "", "invalid or expired session")
		return
	}

	s.logger.Printf("Receiving file %s (session: %s, size: %d)", session.Filename, header.SessionID, header.FileSize)

	// Create a limited reader to read exactly FileSize bytes
	limitedReader := io.LimitReader(conn, header.FileSize)

	// Save file using storage (which handles streaming internally)
	meta, err := s.storage.SaveFile(session.FileID, session.Filename, session.ContentType, limitedReader)
	if err != nil {
		s.logger.Printf("Failed to save file: %v", err)
		s.sendUploadResponse(conn, false, "", fmt.Sprintf("failed to save file: %v", err))
		return
	}

	s.logger.Printf("File saved successfully: %s (hash: %s)", session.Filename, meta.Hash)

	// Remove session after successful upload
	s.sessions.RemoveSession(header.SessionID)

	// Notify master of completed upload
	if s.onUploadComplete != nil {
		s.onUploadComplete(header.SessionID, session.FileID.String(), meta.Hash)
	}

	// Send success response
	s.sendUploadResponse(conn, true, meta.Hash, "")
}

func (s *StreamServer) sendUploadResponse(conn net.Conn, success bool, hash, errMsg string) {
	// Simple response format: 1 byte status + 64 bytes hash + 256 bytes error
	resp := make([]byte, 321)
	if success {
		resp[0] = 1
		copy(resp[1:65], []byte(hash))
	} else {
		resp[0] = 0
		copy(resp[65:321], []byte(errMsg))
	}
	conn.Write(resp)
}

func (s *StreamServer) handleDownload(conn net.Conn) {
	// Read rest of download request (we already read magic byte)
	restBuf := make([]byte, DownloadRequestSize-1)
	if _, err := io.ReadFull(conn, restBuf); err != nil {
		s.logger.Printf("Failed to read download request: %v", err)
		return
	}

	hash := string(trimNull(restBuf[:64]))

	s.logger.Printf("Download request for hash: %s", hash)

	// Get file from storage
	reader, meta, err := s.storage.GetFile(hash)
	if err != nil {
		s.logger.Printf("File not found: %s", hash)
		s.sendDownloadError(conn, fmt.Sprintf("file not found: %v", err))
		return
	}
	defer reader.Close()

	// Send header with file info
	header := &StreamHeader{
		Magic:     MagicDownload,
		SessionID: hash[:36], // Use first 36 chars of hash as session ID
		FileSize:  meta.Size,
		ChunkSize: DefaultChunkSize,
	}

	if err := header.Encode(conn); err != nil {
		s.logger.Printf("Failed to send header: %v", err)
		return
	}

	// Stream file content
	buf := make([]byte, DefaultChunkSize)
	totalSent := int64(0)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			written, writeErr := conn.Write(buf[:n])
			if writeErr != nil {
				s.logger.Printf("Failed to write data: %v", writeErr)
				return
			}
			totalSent += int64(written)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Printf("Failed to read file: %v", err)
			return
		}
	}

	s.logger.Printf("Sent file %s (%d bytes)", hash, totalSent)
}

func (s *StreamServer) sendDownloadError(conn net.Conn, errMsg string) {
	// Send header with size 0 to indicate error, then error message
	header := &StreamHeader{
		Magic:     MagicDownload,
		SessionID: "error",
		FileSize:  -1, // Negative size indicates error
		ChunkSize: 0,
	}
	header.Encode(conn)

	// Send error message (max 256 bytes)
	errBytes := make([]byte, 256)
	copy(errBytes, []byte(errMsg))
	conn.Write(errBytes)
}
