package test

import (
	"testing"

	"video-platform/internal/models"
	"video-platform/internal/services/session"
)

func TestSessionService(t *testing.T) {
	// 创建会话管理服务实例
	service := session.NewService()

	// 测试用例1: 创建会话
	t.Run("CreateSession", func(t *testing.T) {
		streamID := "stream-1"
		clientID := "client-1"

		session, err := service.CreateSession(streamID, clientID)
		if err != nil {
			t.Errorf("CreateSession failed: %v", err)
			return
		}

		if session.ID == "" {
			t.Error("session ID should not be empty")
		}

		if session.StreamID != streamID {
			t.Errorf("expected stream ID %s, got %s", streamID, session.StreamID)
		}

		if session.ClientID != clientID {
			t.Errorf("expected client ID %s, got %s", clientID, session.ClientID)
		}

		if session.Status != models.SessionStatusActive {
			t.Errorf("expected session status %s, got %s", models.SessionStatusActive, session.Status)
		}

		t.Logf("CreateSession test passed, session ID: %s", session.ID)
	})

	// 测试用例2: 获取会话列表
	t.Run("GetSessions", func(t *testing.T) {
		sessions, err := service.GetSessions()
		if err != nil {
			t.Errorf("GetSessions failed: %v", err)
			return
		}

		if len(sessions) == 0 {
			t.Error("session list should not be empty")
		}

		t.Logf("GetSessions test passed, session count: %d", len(sessions))
	})

	// 测试用例3: 关闭会话
	t.Run("CloseSession", func(t *testing.T) {
		// 先创建一个会话
		streamID := "stream-2"
		clientID := "client-2"

		newSession, err := service.CreateSession(streamID, clientID)
		if err != nil {
			t.Errorf("CreateSession failed: %v", err)
			return
		}

		// 关闭会话
		err = service.CloseSession(newSession.ID)
		if err != nil {
			t.Errorf("CloseSession failed: %v", err)
			return
		}

		// 获取会话状态
		status, err := service.GetSessionStatus(newSession.ID)
		if err != nil {
			t.Errorf("GetSessionStatus failed: %v", err)
			return
		}

		if status != models.SessionStatusClosed {
			t.Errorf("expected session status %s, got %s", models.SessionStatusClosed, status)
		}

		t.Logf("CloseSession test passed, session closed: %s", newSession.ID)
	})

	// 测试用例4: 获取会话信息
	t.Run("GetSession", func(t *testing.T) {
		// 先创建一个会话
		streamID := "stream-3"
		clientID := "client-3"

		newSession, err := service.CreateSession(streamID, clientID)
		if err != nil {
			t.Errorf("CreateSession failed: %v", err)
			return
		}

		// 获取会话信息
		session, err := service.GetSession(newSession.ID)
		if err != nil {
			t.Errorf("GetSession failed: %v", err)
			return
		}

		if session.ID != newSession.ID {
			t.Errorf("expected session ID %s, got %s", newSession.ID, session.ID)
		}

		t.Logf("GetSession test passed, session: %s", session.ID)
	})

	// 测试用例5: 获取活跃会话
	t.Run("GetActiveSessions", func(t *testing.T) {
		// 先创建几个会话
		for i := 1; i <= 3; i++ {
			streamID := "stream-" + string(rune('0'+i))
			clientID := "client-" + string(rune('0'+i))

			_, err := service.CreateSession(streamID, clientID)
			if err != nil {
				t.Errorf("CreateSession failed: %v", err)
				return
			}
		}

		// 获取活跃会话
		sessions, err := service.GetActiveSessions()
		if err != nil {
			t.Errorf("GetActiveSessions failed: %v", err)
			return
		}

		if len(sessions) == 0 {
			t.Error("active session list should not be empty")
		}

		t.Logf("GetActiveSessions test passed, active session count: %d", len(sessions))
	})

	// 测试用例6: 清理过期会话
	t.Run("CleanupExpiredSessions", func(t *testing.T) {
		// 清理过期会话
		err := service.CleanupExpiredSessions()
		if err != nil {
			t.Errorf("CleanupExpiredSessions failed: %v", err)
			return
		}

		t.Log("CleanupExpiredSessions test passed")
	})
}
