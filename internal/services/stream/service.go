package stream

import (
	"video-platform/internal/models"
)

// Service 流媒体服务接口
type Service interface {
	// StartStream 启动流
	StartStream(req *models.StartStreamRequest) (*models.Stream, error)
	
	// StopStream 停止流
	StopStream(streamID string) error
	
	// GetStream 获取流信息
	GetStream(streamID string) (*models.Stream, error)
	
	// GetStreams 获取流列表
	GetStreams() ([]models.Stream, error)
	
	// GetStreamStatus 获取流状态
	GetStreamStatus(streamID string) (models.StreamStatus, error)
	
	// UpdateStreamStatus 更新流状态
	UpdateStreamStatus(streamID string, status models.StreamStatus) error
	
	// GetStreamURL 获取流地址
	GetStreamURL(streamID string) (string, error)
	
	// RestartStream 重启流
	RestartStream(streamID string) error
	
	// CreateStream 创建流
	CreateStream(stream *models.Stream) error
	
	// DeleteStream 删除流
	DeleteStream(streamID string) error
}


