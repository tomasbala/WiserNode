package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wiserstream/gb28181-server/internal/services/device"
)

type DeviceHandler struct {
	deviceSvc *device.DeviceService
}

func NewDeviceHandler(deviceSvc *device.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceSvc: deviceSvc}
}

func (h *DeviceHandler) ListDevices(c *gin.Context) {
	devices := h.deviceSvc.ListDevices()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":   len(devices),
			"devices": devices,
		},
	})
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("device_id")

	device, exists := h.deviceSvc.GetDevice(deviceID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "device not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": device,
	})
}

func (h *DeviceHandler) GetChannels(c *gin.Context) {
	deviceID := c.Param("device_id")

	channels, exists := h.deviceSvc.GetChannels(deviceID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "device not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":    len(channels),
			"channels": channels,
		},
	})
}
