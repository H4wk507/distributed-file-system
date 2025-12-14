package client

import (
	"dfs-backend/dfs/common"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

type MasterClient struct {
	masterIP   string
	masterPort int
	timeout    time.Duration
	clientID   uuid.UUID
}

func NewMasterClient(ip string, port int) *MasterClient {
	return &MasterClient{
		masterIP:   ip,
		masterPort: port,
		timeout:    30 * time.Second,
		clientID:   uuid.New(),
	}
}

func (c *MasterClient) UploadFile(fileID uuid.UUID, filename, contentType string, data []byte) (*common.APIFileUploadResponse, error) {
	requestID := uuid.New().String()

	request := common.APIFileUploadRequest{
		RequestID:   requestID,
		FileID:      fileID,
		Filename:    filename,
		ContentType: contentType,
		Size:        int64(len(data)),
		Data:        data,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal upload request: %w", err)
	}

	msg := common.Message{
		Type:    common.MessageAPIFileUpload,
		From:    c.clientID,
		Payload: payload,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send upload request: %w", err)
	}

	var uploadResponse common.APIFileUploadResponse
	if err := json.Unmarshal(response.Payload, &uploadResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}

	return &uploadResponse, nil
}

func (c *MasterClient) DownloadFile(fileID uuid.UUID, hash string) (*common.APIFileDownloadResponse, error) {
	requestID := uuid.New().String()

	request := common.APIFileDownloadRequest{
		RequestID: requestID,
		FileID:    fileID,
		Hash:      hash,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal download request: %w", err)
	}

	msg := common.Message{
		Type:    common.MessageAPIFileDownload,
		From:    c.clientID,
		Payload: payload,
	}

	response, err := c.sendRequest(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send download request: %w", err)
	}

	var downloadResponse common.APIFileDownloadResponse
	if err := json.Unmarshal(response.Payload, &downloadResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal download response: %w", err)
	}

	return &downloadResponse, nil
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

func (c *MasterClient) sendRequest(msg common.Message) (*common.Message, error) {
	addr := fmt.Sprintf("%s:%d", c.masterIP, c.masterPort)

	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master at %s: %w", addr, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	msgWithTime := common.MessageWithTime{
		Message:     msg,
		LogicalTime: 0, // API client doesn't participate in logical clock
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msgWithTime); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	decoder := json.NewDecoder(conn)
	var responseWithTime common.MessageWithTime
	if err := decoder.Decode(&responseWithTime); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &responseWithTime.Message, nil
}

func (c *MasterClient) Ping() error {
	addr := fmt.Sprintf("%s:%d", c.masterIP, c.masterPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("master not reachable at %s: %w", addr, err)
	}
	conn.Close()
	return nil
}
