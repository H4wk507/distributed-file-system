package streaming

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	SessionTimeout = 5 * time.Minute
)

type UploadSession struct {
	SessionID   string
	FileID      uuid.UUID
	Filename    string
	ContentType string
	Size        int64
	CreatedAt   time.Time
}

type SessionManager struct {
	sessions map[string]*UploadSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*UploadSession),
	}
}

func (m *SessionManager) CreateSession(fileID uuid.UUID, filename, contentType string, size int64) *UploadSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &UploadSession{
		SessionID:   uuid.New().String(),
		FileID:      fileID,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		CreatedAt:   time.Now(),
	}

	m.sessions[session.SessionID] = session
	return session
}

func (m *SessionManager) GetSession(sessionID string) (*UploadSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Since(session.CreatedAt) > SessionTimeout {
		return nil, false
	}

	return session, true
}

func (m *SessionManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

func (m *SessionManager) AddSession(session *UploadSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.SessionID] = session
}

func (m *SessionManager) CleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.Sub(session.CreatedAt) > SessionTimeout {
			delete(m.sessions, id)
		}
	}
}

func (m *SessionManager) StartCleanupRoutine(stopChan <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.CleanupExpired()
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}
