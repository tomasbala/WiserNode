package models

import (
	"time"
)

// DeviceType 设备类型
type DeviceType string

const (
	DeviceTypeONVIF    DeviceType = "onvif"
	DeviceTypeGB28181  DeviceType = "gb28181"
	DeviceTypeRTSP     DeviceType = "rtsp"
	DeviceTypeUnknown  DeviceType = "unknown"
)

// DeviceStatus 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusOffline  DeviceStatus = "offline"
	DeviceStatusError    DeviceStatus = "error"
	DeviceStatusUnknown  DeviceStatus = "unknown"
)

// StreamStatus 流状态
type StreamStatus string

const (
	StreamStatusPlaying  StreamStatus = "playing"
	StreamStatusStopped  StreamStatus = "stopped"
	StreamStatusError    StreamStatus = "error"
	StreamStatusUnknown  StreamStatus = "unknown"
)

// SessionStatus 会话状态
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusClosed  SessionStatus = "closed"
	SessionStatusError   SessionStatus = "error"
	SessionStatusUnknown SessionStatus = "unknown"
)

// Device 设备模型
type Device struct {
	ID           string      `json:"id" db:"id"`
	Name         string      `json:"name" db:"name"`
	Type         DeviceType  `json:"type" db:"type"`
	IP           string      `json:"ip" db:"ip"`
	Port         int         `json:"port" db:"port"`
	Username     string      `json:"username" db:"username"`
	Password     string      `json:"password" db:"password"`
	Status       DeviceStatus `json:"status" db:"status"`
	Manufacturer string      `json:"manufacturer" db:"manufacturer"`
	Model        string      `json:"model" db:"model"`
	Firmware     string      `json:"firmware" db:"firmware"`
	SerialNumber string      `json:"serial_number" db:"serial_number"`
	Channels     []Channel   `json:"channels" db:"-"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// Channel 通道模型
type Channel struct {
	ID           string    `json:"id" db:"id"`
	DeviceID     string    `json:"device_id" db:"device_id"`
	Name         string    `json:"name" db:"name"`
	ChannelID    string    `json:"channel_id" db:"channel_id"`
	Status       string    `json:"status" db:"status"`
	StreamURL    string    `json:"stream_url" db:"stream_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Stream 流模型
type Stream struct {
	ID         string       `json:"id" db:"id"`
	DeviceID   string       `json:"device_id" db:"device_id"`
	ChannelID  string       `json:"channel_id" db:"channel_id"`
	Name       string       `json:"name" db:"name"`
	URL        string       `json:"url" db:"url"`
	Status     StreamStatus `json:"status" db:"status"`
	Protocol   string       `json:"protocol" db:"protocol"`
	Codec      string       `json:"codec" db:"codec"`
	Resolution string       `json:"resolution" db:"resolution"`
	Bitrate    int          `json:"bitrate" db:"bitrate"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
}

// Session 会话模型
type Session struct {
	ID         string        `json:"id" db:"id"`
	StreamID   string        `json:"stream_id" db:"stream_id"`
	ClientID   string        `json:"client_id" db:"client_id"`
	Status     SessionStatus `json:"status" db:"status"`
	StartTime  time.Time     `json:"start_time" db:"start_time"`
	EndTime    time.Time     `json:"end_time" db:"end_time"`
	CreatedAt  time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at" db:"updated_at"`
}

// DiscoverDevicesRequest 发现设备请求
type DiscoverDevicesRequest struct {
	Timeout int `json:"timeout"`
}

// DiscoverDevicesResponse 发现设备响应
type DiscoverDevicesResponse struct {
	Devices []Device `json:"devices"`
}

// AddDeviceRequest 添加设备请求
type AddDeviceRequest struct {
	Name     string     `json:"name" binding:"required"`
	Type     DeviceType `json:"type" binding:"required"`
	IP       string     `json:"ip" binding:"required,ip"`
	Port     int        `json:"port"`
	Username string     `json:"username"`
	Password string     `json:"password"`
}

// AddDeviceResponse 添加设备响应
type AddDeviceResponse struct {
	Device Device `json:"device"`
}

// GetDeviceStatusRequest 获取设备状态请求
type GetDeviceStatusRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
}

// GetDeviceStatusResponse 获取设备状态响应
type GetDeviceStatusResponse struct {
	Status DeviceStatus `json:"status"`
}

// StartStreamRequest 启动流请求
type StartStreamRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	ChannelID string `json:"channel_id"`
	Protocol  string `json:"protocol"`
}

// StartStreamResponse 启动流响应
type StartStreamResponse struct {
	Stream Stream `json:"stream"`
}

// StopStreamRequest 停止流请求
type StopStreamRequest struct {
	StreamID string `json:"stream_id" binding:"required"`
}

// StopStreamResponse 停止流响应
type StopStreamResponse struct {
	Success bool `json:"success"`
}

// GetStreamUrlRequest 获取流地址请求
type GetStreamUrlRequest struct {
	StreamID string `json:"stream_id" binding:"required"`
}

// GetStreamUrlResponse 获取流地址响应
type GetStreamUrlResponse struct {
	URL string `json:"url"`
}
