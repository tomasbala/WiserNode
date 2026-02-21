package device

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"video-platform/internal/models"
)

// 实现具体的设备管理服务逻辑

// 设备管理服务实现
type deviceService struct {
	devices     map[string]*models.Device
	channels    map[string][]models.Channel
	mutex       sync.RWMutex
	nextDeviceID int
	nextChannelID int
}

// NewService 创建设备管理服务实例
func NewService() Service {
	return &deviceService{
		devices:     make(map[string]*models.Device),
		channels:    make(map[string][]models.Channel),
		nextDeviceID: 1,
		nextChannelID: 1,
	}
}

// DiscoverDevices 发现设备
func (s *deviceService) DiscoverDevices(timeout int) ([]models.Device, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 模拟设备发现
	// 实际实现中应该：
	// 1. 发送广播包
	// 2. 接收设备响应
	// 3. 解析设备信息

	devices := make([]models.Device, 0, len(s.devices))
	for _, d := range s.devices {
		devices = append(devices, *d)
	}

	return devices, nil
}

// AddDevice 添加设备
func (s *deviceService) AddDevice(req *models.AddDeviceRequest) (*models.Device, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证设备信息
	if req.Name == "" {
		return nil, errors.New("device name is required")
	}

	if req.IP == "" {
		return nil, errors.New("device IP is required")
	}

	// 生成设备ID
	deviceID := fmt.Sprintf("device-%d", s.nextDeviceID)
	s.nextDeviceID++

	// 创建设备
	device := &models.Device{
		ID:           deviceID,
		Name:         req.Name,
		Type:         req.Type,
		IP:           req.IP,
		Port:         req.Port,
		Username:     req.Username,
		Password:     req.Password,
		Status:       models.DeviceStatusOnline,
		Manufacturer: "Unknown",
		Model:        "Unknown",
		Firmware:     "Unknown",
		SerialNumber: "Unknown",
		Channels:     []models.Channel{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存设备
	s.devices[deviceID] = device

	// 模拟添加通道
	channels := []models.Channel{
		{
			ID:           fmt.Sprintf("channel-%d", s.nextChannelID),
			DeviceID:     deviceID,
			Name:         "Channel 1",
			ChannelID:    "1",
			Status:       "online",
			StreamURL:    fmt.Sprintf("rtsp://%s:554/stream/1", req.IP),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	s.nextChannelID++

	s.channels[deviceID] = channels
	device.Channels = channels

	return device, nil
}

// GetDevice 获取设备信息
func (s *deviceService) GetDevice(deviceID string) (*models.Device, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return nil, errors.New("device not found")
	}

	// 更新通道信息
	device.Channels = s.channels[deviceID]

	return device, nil
}

// GetDevices 获取设备列表
func (s *deviceService) GetDevices() ([]models.Device, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	devices := make([]models.Device, 0, len(s.devices))
	for _, d := range s.devices {
		// 更新通道信息
		d.Channels = s.channels[d.ID]
		devices = append(devices, *d)
	}

	return devices, nil
}

// UpdateDevice 更新设备信息
func (s *deviceService) UpdateDevice(device *models.Device) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	existingDevice, exists := s.devices[device.ID]
	if !exists {
		return errors.New("device not found")
	}

	// 更新设备信息
	existingDevice.Name = device.Name
	existingDevice.IP = device.IP
	existingDevice.Port = device.Port
	existingDevice.Username = device.Username
	existingDevice.Password = device.Password
	existingDevice.Manufacturer = device.Manufacturer
	existingDevice.Model = device.Model
	existingDevice.Firmware = device.Firmware
	existingDevice.SerialNumber = device.SerialNumber
	existingDevice.UpdatedAt = time.Now()

	return nil
}

// DeleteDevice 删除设备
func (s *deviceService) DeleteDevice(deviceID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查设备是否存在
	_, exists := s.devices[deviceID]
	if !exists {
		return errors.New("device not found")
	}

	// 删除设备
	delete(s.devices, deviceID)
	delete(s.channels, deviceID)

	return nil
}

// GetDeviceStatus 获取设备状态
func (s *deviceService) GetDeviceStatus(deviceID string) (models.DeviceStatus, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return models.DeviceStatusUnknown, errors.New("device not found")
	}

	return device.Status, nil
}

// UpdateDeviceStatus 更新设备状态
func (s *deviceService) UpdateDeviceStatus(deviceID string, status models.DeviceStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	device, exists := s.devices[deviceID]
	if !exists {
		return errors.New("device not found")
	}

	device.Status = status
	device.UpdatedAt = time.Now()

	return nil
}

// GetDeviceChannels 获取设备通道
func (s *deviceService) GetDeviceChannels(deviceID string) ([]models.Channel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 检查设备是否存在
	_, exists := s.devices[deviceID]
	if !exists {
		return nil, errors.New("device not found")
	}

	channels, exists := s.channels[deviceID]
	if !exists {
		return []models.Channel{}, nil
	}

	return channels, nil
}
