package stream

import (
	"fmt"
	"sync"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/models"
)

type StreamService struct {
	zlm     *ZLMClient
	streams map[string]*models.Stream
	mu      sync.RWMutex
}

func NewStreamService(cfg *config.Config) *StreamService {
	return &StreamService{
		zlm:     NewZLMClient(&cfg.ZLMediaKit),
		streams: make(map[string]*models.Stream),
	}
}

func (s *StreamService) OpenStream(deviceID, channelID, ssrc string) (*models.Stream, error) {
	streamID := fmt.Sprintf("%s_%s", deviceID, channelID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if stream, exists := s.streams[streamID]; exists {
		return stream, nil
	}

	port, err := s.zlm.GetRTPPort(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rtp port: %w", err)
	}

	stream := &models.Stream{
		StreamID:   streamID,
		DeviceID:   deviceID,
		ChannelID:  channelID,
		SSRC:       ssrc,
		RTPPort:    port,
		RTSPUrl:    s.zlm.GetStreamURL(streamID, "rtsp"),
		FLVUrl:     s.zlm.GetStreamURL(streamID, "flv"),
		WebRTCUrl:  s.zlm.GetStreamURL(streamID, "webrtc"),
		HLSUrl:     s.zlm.GetStreamURL(streamID, "hls"),
		Status:     "waiting",
		CreateTime: time.Now(),
	}

	s.streams[streamID] = stream
	return stream, nil
}

func (s *StreamService) CloseStream(streamID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.streams[streamID]; !exists {
		return nil
	}

	_, err := s.zlm.CloseRTPServer(streamID)
	if err != nil {
		return fmt.Errorf("failed to close rtp server: %w", err)
	}

	delete(s.streams, streamID)
	return nil
}

func (s *StreamService) GetStream(streamID string) (*models.Stream, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stream, exists := s.streams[streamID]
	return stream, exists
}

func (s *StreamService) ListStreams() []*models.Stream {
	s.mu.RLock()
	defer s.mu.RUnlock()

	streams := make([]*models.Stream, 0, len(s.streams))
	for _, stream := range s.streams {
		streams = append(streams, stream)
	}
	return streams
}

func (s *StreamService) UpdateStreamStatus(streamID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stream, exists := s.streams[streamID]; exists {
		stream.Status = status
	}
}

func (s *StreamService) GetZLMClient() *ZLMClient {
	return s.zlm
}
