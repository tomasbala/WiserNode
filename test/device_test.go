package test

import (
	"testing"

	"video-platform/internal/models"
	"video-platform/internal/services/device"
)

func TestDeviceService(t *testing.T) {
	// 创建设备管理服务实例
	service := device.NewService()

	// 测试用例1: 添加设备
	t.Run("AddDevice", func(t *testing.T) {
		req := &models.AddDeviceRequest{
			Name:     "测试摄像头",
			Type:     models.DeviceTypeONVIF,
			IP:       "192.168.1.100",
			Port:     80,
			Username: "admin",
			Password: "admin123",
		}

		device, err := service.AddDevice(req)
		if err != nil {
			t.Errorf("AddDevice failed: %v", err)
			return
		}

		if device.ID == "" {
			t.Error("device ID should not be empty")
		}

		if device.Name != req.Name {
			t.Errorf("expected device name %s, got %s", req.Name, device.Name)
		}

		t.Logf("AddDevice test passed, device ID: %s", device.ID)
	})

	// 测试用例2: 获取设备列表
	t.Run("GetDevices", func(t *testing.T) {
		devices, err := service.GetDevices()
		if err != nil {
			t.Errorf("GetDevices failed: %v", err)
			return
		}

		if len(devices) == 0 {
			t.Error("device list should not be empty")
		}

		t.Logf("GetDevices test passed, device count: %d", len(devices))
	})

	// 测试用例3: 获取设备信息
	t.Run("GetDevice", func(t *testing.T) {
		// 先添加一个设备
		req := &models.AddDeviceRequest{
			Name:     "测试摄像头2",
			Type:     models.DeviceTypeRTSP,
			IP:       "192.168.1.101",
			Port:     554,
		}

		newDevice, err := service.AddDevice(req)
		if err != nil {
			t.Errorf("AddDevice failed: %v", err)
			return
		}

		// 获取设备信息
		device, err := service.GetDevice(newDevice.ID)
		if err != nil {
			t.Errorf("GetDevice failed: %v", err)
			return
		}

		if device.ID != newDevice.ID {
			t.Errorf("expected device ID %s, got %s", newDevice.ID, device.ID)
		}

		t.Logf("GetDevice test passed, device: %s", device.Name)
	})

	// 测试用例4: 删除设备
	t.Run("DeleteDevice", func(t *testing.T) {
		// 先添加一个设备
		req := &models.AddDeviceRequest{
			Name:     "测试摄像头3",
			Type:     models.DeviceTypeGB28181,
			IP:       "192.168.1.102",
			Port:     5060,
		}

		newDevice, err := service.AddDevice(req)
		if err != nil {
			t.Errorf("AddDevice failed: %v", err)
			return
		}

		// 删除设备
		err = service.DeleteDevice(newDevice.ID)
		if err != nil {
			t.Errorf("DeleteDevice failed: %v", err)
			return
		}

		// 验证设备已删除
		_, err = service.GetDevice(newDevice.ID)
		if err == nil {
			t.Error("device should be deleted")
		}

		t.Logf("DeleteDevice test passed, device deleted: %s", newDevice.ID)
	})

	// 测试用例5: 获取设备状态
	t.Run("GetDeviceStatus", func(t *testing.T) {
		// 先添加一个设备
		req := &models.AddDeviceRequest{
			Name:     "测试摄像头4",
			Type:     models.DeviceTypeONVIF,
			IP:       "192.168.1.103",
		}

		newDevice, err := service.AddDevice(req)
		if err != nil {
			t.Errorf("AddDevice failed: %v", err)
			return
		}

		// 获取设备状态
		status, err := service.GetDeviceStatus(newDevice.ID)
		if err != nil {
			t.Errorf("GetDeviceStatus failed: %v", err)
			return
		}

		if status == models.DeviceStatusUnknown {
			t.Error("device status should not be unknown")
		}

		t.Logf("GetDeviceStatus test passed, status: %s", status)
	})

	// 测试用例6: 获取设备通道
	t.Run("GetDeviceChannels", func(t *testing.T) {
		// 先添加一个设备
		req := &models.AddDeviceRequest{
			Name:     "测试摄像头5",
			Type:     models.DeviceTypeONVIF,
			IP:       "192.168.1.104",
		}

		newDevice, err := service.AddDevice(req)
		if err != nil {
			t.Errorf("AddDevice failed: %v", err)
			return
		}

		// 获取设备通道
		channels, err := service.GetDeviceChannels(newDevice.ID)
		if err != nil {
			t.Errorf("GetDeviceChannels failed: %v", err)
			return
		}

		if len(channels) == 0 {
			t.Error("device channels should not be empty")
		}

		t.Logf("GetDeviceChannels test passed, channel count: %d", len(channels))
	})
}
