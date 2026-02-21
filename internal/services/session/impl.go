package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"video-platform/internal/models"
)

// 实现具体的会话管理服务逻辑

// 会话管理服务实现
type sessionService struct {
	sessions     map[string]*models.Session
	mutex       sync.RWMutex
	nextSessionID int
}

// NewService 创建会话管理服务实例
func NewService() Service {
	return &sessionService{
		sessions:     make(map[string]*models.Session),
		nextSessionID: 1,
	}
}

// CreateSession 创建会话
func (s *sessionService) CreateSession(streamID string, clientID string) (*models.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证请求
	if streamID == "" {
		return nil, errors.New("stream ID is required")
	}

	if clientID == "" {
		return nil, errors.New("client ID is required")
	}

	// 生成会话ID
	sessionID := fmt.Sprintf("session-%d", s.nextSessionID)
	s.nextSessionID++

	// 创建会话
	session := &models.Session{
		ID:         sessionID,
		StreamID:   streamID,
		ClientID:   clientID,
		Status:     models.SessionStatusActive,
		StartTime:  time.Now(),
		EndTime:    time.Time{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存会话
	s.sessions[sessionID] = session

	return session, nil
}

// CloseSession 关闭会话
func (s *sessionService) CloseSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	// 更新会话状态
	session.Status = models.SessionStatusClosed
	session.EndTime = time.Now()
	session.UpdatedAt = time.Now()

	return nil
}

// GetSession 获取会话信息
func (s *sessionService) GetSession(sessionID string) (*models.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	return session, nil
}

// GetSessions 获取会话列表
func (s *sessionService) GetSessions() ([]models.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sessions := make([]models.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

// GetSessionStatus 获取会话状态
func (s *sessionService) GetSessionStatus(sessionID string) (models.SessionStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return models.SessionStatusUnknown, errors.New("session not found")
	}

	return session.Status, nil
}

// UpdateSessionStatus 更新会话状态
func (s *sessionService) UpdateSessionStatus(sessionID string, status models.SessionStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	session.Status = status
	session.UpdatedAt = time.Now()

	if status == models.SessionStatusClosed && session.EndTime.IsZero() {
		session.EndTime = time.Now()
	}

	return nil
}

// GetActiveSessions 获取活跃会话
func (s *sessionService) GetActiveSessions() ([]models.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var activeSessions []models.Session
	for _, session := range s.sessions {
		if session.Status == models.SessionStatusActive {
			activeSessions = append(activeSessions, *session)
		}
	}

	return activeSessions, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *sessionService) CleanupExpiredSessions() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 清理超过24小时的已关闭会话
	expiryTime := time.Now().Add(-24 * time.Hour)
	for sessionID, session := range s.sessions {
		if session.Status == models.SessionStatusClosed && session.EndTime.Before(expiryTime) {
			delete(s.sessions, sessionID)
		}
	}

	return nil
}
