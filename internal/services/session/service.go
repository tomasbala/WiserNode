package session

import (
	"video-platform/internal/models"
)

// Service 会话管理服务接口
type Service interface {
	// CreateSession 创建会话
	CreateSession(streamID string, clientID string) (*models.Session, error)
	
	// CloseSession 关闭会话
	CloseSession(sessionID string) error
	
	// GetSession 获取会话信息
	GetSession(sessionID string) (*models.Session, error)
	
	// GetSessions 获取会话列表
	GetSessions() ([]models.Session, error)
	
	// GetSessionStatus 获取会话状态
	GetSessionStatus(sessionID string) (models.SessionStatus, error)
	
	// UpdateSessionStatus 更新会话状态
	UpdateSessionStatus(sessionID string, status models.SessionStatus) error
	
	// GetActiveSessions 获取活跃会话
	GetActiveSessions() ([]models.Session, error)
	
	// CleanupExpiredSessions 清理过期会话
	CleanupExpiredSessions() error
}


