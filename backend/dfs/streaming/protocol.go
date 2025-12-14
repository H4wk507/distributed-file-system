package streaming

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const (
	// HeaderSize is the fixed size of the binary stream header
	HeaderSize = 64
	// DefaultChunkSize for streaming data (4MB)
	DefaultChunkSize = 4 * 1024 * 1024
	// StreamPortOffset is added to the node's main port
	StreamPortOffset = 100
	// Magic bytes to identify stream protocol
	MagicUpload   byte = 0x01
	MagicDownload byte = 0x02
)

var BufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, DefaultChunkSize)
		return &buf
	},
}

func GetBuffer() *[]byte {
	return BufPool.Get().(*[]byte)
}

func PutBuffer(buf *[]byte) {
	BufPool.Put(buf)
}

// StreamHeader is the binary header for file streaming
// Total size: 64 bytes (fixed)
//
// Layout:
//
//	[0]      - Magic byte (0x01 = upload, 0x02 = download)
//	[1:37]   - SessionID (36 bytes, UUID string)
//	[37:45]  - FileSize (8 bytes, int64, big endian)
//	[45:49]  - ChunkSize (4 bytes, int32, big endian)
//	[49:64]  - Reserved/padding (15 bytes)
type StreamHeader struct {
	Magic     byte
	SessionID string
	FileSize  int64
	ChunkSize int32
}

func (h *StreamHeader) Encode(w io.Writer) error {
	buf := make([]byte, HeaderSize)

	buf[0] = h.Magic

	// SessionID (36 bytes)
	sessionBytes := []byte(h.SessionID)
	if len(sessionBytes) > 36 {
		sessionBytes = sessionBytes[:36]
	}
	copy(buf[1:37], sessionBytes)

	// FileSize (8 bytes, big endian)
	binary.BigEndian.PutUint64(buf[37:45], uint64(h.FileSize))

	// ChunkSize (4 bytes, big endian)
	binary.BigEndian.PutUint32(buf[45:49], uint32(h.ChunkSize))

	// Remaining bytes are padding (already zero)

	_, err := w.Write(buf)
	return err
}

func DecodeHeader(r io.Reader) (*StreamHeader, error) {
	buf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	header := &StreamHeader{
		Magic:     buf[0],
		SessionID: string(trimNull(buf[1:37])),
		FileSize:  int64(binary.BigEndian.Uint64(buf[37:45])),
		ChunkSize: int32(binary.BigEndian.Uint32(buf[45:49])),
	}

	if header.Magic != MagicUpload && header.Magic != MagicDownload {
		return nil, fmt.Errorf("invalid magic byte: %x", header.Magic)
	}

	return header, nil
}

func trimNull(b []byte) []byte {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			return b[:i+1]
		}
	}
	return b[:0]
}

// DownloadRequest is sent by client to request a file download
// Layout (64 bytes):
//
//	[0]      - Magic byte (0x02)
//	[1:65]   - Hash (64 bytes, SHA256 hex string)
type DownloadRequest struct {
	Hash string
}

const DownloadRequestSize = 65

func (r *DownloadRequest) Encode(w io.Writer) error {
	buf := make([]byte, DownloadRequestSize)
	buf[0] = MagicDownload
	copy(buf[1:65], []byte(r.Hash))
	_, err := w.Write(buf)
	return err
}

func DecodeDownloadRequest(r io.Reader) (*DownloadRequest, error) {
	buf := make([]byte, DownloadRequestSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("failed to read download request: %w", err)
	}

	if buf[0] != MagicDownload {
		return nil, fmt.Errorf("invalid magic byte for download: %x", buf[0])
	}

	return &DownloadRequest{
		Hash: string(trimNull(buf[1:65])),
	}, nil
}
