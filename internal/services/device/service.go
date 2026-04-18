package device

import (
	"sync"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/models"
)

type DeviceService struct {
	cfg     *config.Config
	devices map[string]*models.Device
	mu      sync.RWMutex
}

func NewDeviceService(cfg *config.Config) *DeviceService {
	svc := &DeviceService{
		cfg:     cfg,
		devices: make(map[string]*models.Device),
	}

	go svc.heartbeatCheckLoop()

	return svc
}

func (s *DeviceService) RegisterDevice(deviceID, ip string, port int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if device, exists := s.devices[deviceID]; exists {
		device.IP = ip
		device.Port = port
		device.RegisterTime = now
		device.KeepaliveTime = now
		device.Status = models.DeviceOnline
		return
	}

	s.devices[deviceID] = &models.Device{
		DeviceID:      deviceID,
		IP:            ip,
		Port:          port,
		RegisterTime:  now,
		KeepaliveTime: now,
		Status:        models.DeviceOnline,
		Channels:      make([]models.Channel, 0),
	}
}

func (s *DeviceService) UnregisterDevice(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if device, exists := s.devices[deviceID]; exists {
		device.Status = models.DeviceOffline
	}
}

func (s *DeviceService) UpdateKeepalive(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if device, exists := s.devices[deviceID]; exists {
		device.KeepaliveTime = time.Now()
		device.Status = models.DeviceOnline
	}
}

func (s *DeviceService) UpdateChannels(deviceID string, channels []models.Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if device, exists := s.devices[deviceID]; exists {
		device.Channels = channels
	}
}

func (s *DeviceService) GetDevice(deviceID string) (*models.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return nil, false
	}
	return device, true
}

func (s *DeviceService) ListDevices() []*models.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*models.Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices
}

func (s *DeviceService) GetChannels(deviceID string) ([]models.Channel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return nil, false
	}
	return device.Channels, true
}

func (s *DeviceService) heartbeatCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkDeviceHeartbeat()
	}
}

func (s *DeviceService) checkDeviceHeartbeat() {
	s.mu.Lock()
	defer s.mu.Unlock()

	timeout := time.Duration(s.cfg.Device.HeartbeatTimeout) * time.Second
	now := time.Now()

	for _, device := range s.devices {
		if device.Status == models.DeviceOnline {
			if now.Sub(device.KeepaliveTime) > timeout {
				device.Status = models.DeviceOffline
			}
		}
	}
}
