package stream

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"video-platform/internal/models"
)

// 实现具体的流媒体服务逻辑

// 流媒体服务实现
type streamService struct {
	streams     map[string]*models.Stream
	mutex       sync.RWMutex
	nextStreamID int
}

// NewService 创建流媒体服务实例
func NewService() Service {
	return &streamService{
		streams:     make(map[string]*models.Stream),
		nextStreamID: 1,
	}
}

// StartStream 启动流
func (s *streamService) StartStream(req *models.StartStreamRequest) (*models.Stream, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证请求
	if req.DeviceID == "" {
		return nil, errors.New("device ID is required")
	}

	// 生成流ID
	streamID := fmt.Sprintf("stream-%d", s.nextStreamID)
	s.nextStreamID++

	// 构建流URL
	streamURL := fmt.Sprintf("rtsp://%s:554/stream/%s", req.DeviceID, req.ChannelID)
	if req.ChannelID == "" {
		streamURL = fmt.Sprintf("rtsp://%s:554/stream/1", req.DeviceID)
	}

	// 创建流
	stream := &models.Stream{
		ID:         streamID,
		DeviceID:   req.DeviceID,
		ChannelID:  req.ChannelID,
		Name:       fmt.Sprintf("Stream for device %s", req.DeviceID),
		URL:        streamURL,
		Status:     models.StreamStatusPlaying,
		Protocol:   req.Protocol,
		Codec:      "H.264",
		Resolution: "1920x1080",
		Bitrate:    2048,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存流
	s.streams[streamID] = stream

	return stream, nil
}

// StopStream 停止流
func (s *streamService) StopStream(streamID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	// 更新流状态
	stream.Status = models.StreamStatusStopped
	stream.UpdatedAt = time.Now()

	return nil
}

// GetStream 获取流信息
func (s *streamService) GetStream(streamID string) (*models.Stream, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return nil, errors.New("stream not found")
	}

	return stream, nil
}

// GetStreams 获取流列表
func (s *streamService) GetStreams() ([]models.Stream, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	streams := make([]models.Stream, 0, len(s.streams))
	for _, stream := range s.streams {
		streams = append(streams, *stream)
	}

	return streams, nil
}

// GetStreamStatus 获取流状态
func (s *streamService) GetStreamStatus(streamID string) (models.StreamStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return models.StreamStatusUnknown, errors.New("stream not found")
	}

	return stream.Status, nil
}

// UpdateStreamStatus 更新流状态
func (s *streamService) UpdateStreamStatus(streamID string, status models.StreamStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	stream.Status = status
	stream.UpdatedAt = time.Now()

	return nil
}

// GetStreamURL 获取流地址
func (s *streamService) GetStreamURL(streamID string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return "", errors.New("stream not found")
	}

	return stream.URL, nil
}

// RestartStream 重启流
func (s *streamService) RestartStream(streamID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stream, exists := s.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	// 先停止流
	stream.Status = models.StreamStatusStopped
	stream.UpdatedAt = time.Now()

	// 再启动流
	stream.Status = models.StreamStatusPlaying
	stream.UpdatedAt = time.Now()

	return nil
}

// CreateStream 创建流
func (s *streamService) CreateStream(stream *models.Stream) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证流信息
	if stream.ID == "" {
		stream.ID = fmt.Sprintf("stream-%d", s.nextStreamID)
		s.nextStreamID++
	}

	if stream.Name == "" {
		stream.Name = fmt.Sprintf("Stream %s", stream.ID)
	}

	if stream.Status == "" {
		stream.Status = models.StreamStatusStopped
	}

	if stream.CreatedAt.IsZero() {
		stream.CreatedAt = time.Now()
	}

	stream.UpdatedAt = time.Now()

	// 保存流
	s.streams[stream.ID] = stream

	return nil
}

// DeleteStream 删除流
func (s *streamService) DeleteStream(streamID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查流是否存在
	_, exists := s.streams[streamID]
	if !exists {
		return errors.New("stream not found")
	}

	// 删除流
	delete(s.streams, streamID)

	return nil
}
