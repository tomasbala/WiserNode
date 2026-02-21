package device

import (
	"video-platform/internal/models"
)

// Service 设备管理服务接口
type Service interface {
	// DiscoverDevices 发现设备
	DiscoverDevices(timeout int) ([]models.Device, error)
	
	// AddDevice 添加设备
	AddDevice(req *models.AddDeviceRequest) (*models.Device, error)
	
	// GetDevice 获取设备信息
	GetDevice(deviceID string) (*models.Device, error)
	
	// GetDevices 获取设备列表
	GetDevices() ([]models.Device, error)
	
	// UpdateDevice 更新设备信息
	UpdateDevice(device *models.Device) error
	
	// DeleteDevice 删除设备
	DeleteDevice(deviceID string) error
	
	// GetDeviceStatus 获取设备状态
	GetDeviceStatus(deviceID string) (models.DeviceStatus, error)
	
	// UpdateDeviceStatus 更新设备状态
	UpdateDeviceStatus(deviceID string, status models.DeviceStatus) error
	
	// GetDeviceChannels 获取设备通道
	GetDeviceChannels(deviceID string) ([]models.Channel, error)
}


