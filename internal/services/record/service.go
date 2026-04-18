package record

import (
	"sync"
	"time"

	"github.com/wiserstream/gb28181-server/internal/models"
)

type RecordService struct {
	records  map[string][]models.RecordItem
	playback map[string]*models.PlayBack
	mu       sync.RWMutex
}

func NewRecordService() *RecordService {
	return &RecordService{
		records:  make(map[string][]models.RecordItem),
		playback: make(map[string]*models.PlayBack),
	}
}

func (s *RecordService) UpdateRecords(channelID string, records []models.RecordItem) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[channelID] = records
}

func (s *RecordService) GetRecords(channelID string) []models.RecordItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, exists := s.records[channelID]
	if !exists {
		return nil
	}

	result := make([]models.RecordItem, len(records))
	copy(result, records)
	return result
}

func (s *RecordService) ClearRecords(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.records, channelID)
}

func (s *RecordService) CreatePlayback(deviceID, channelID, ssrc string, rtpPort int, startTime, endTime string) *models.PlayBack {
	s.mu.Lock()
	defer s.mu.Unlock()

	streamID := "playback_" + deviceID + "_" + channelID

	pb := &models.PlayBack{
		StreamID:   streamID,
		DeviceID:   deviceID,
		ChannelID:  channelID,
		StartTime:  startTime,
		EndTime:    endTime,
		SSRC:       ssrc,
		RTPPort:    rtpPort,
		Status:     "waiting",
		CreateTime: time.Now(),
	}

	s.playback[streamID] = pb
	return pb
}

func (s *RecordService) GetPlayback(streamID string) (*models.PlayBack, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pb, exists := s.playback[streamID]
	return pb, exists
}

func (s *RecordService) UpdatePlaybackStatus(streamID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pb, exists := s.playback[streamID]; exists {
		pb.Status = status
	}
}

func (s *RecordService) RemovePlayback(streamID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.playback, streamID)
}

func (s *RecordService) ListPlaybacks() []*models.PlayBack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.PlayBack, 0, len(s.playback))
	for _, pb := range s.playback {
		result = append(result, pb)
	}
	return result
}
