package models

import "time"

type Device struct {
	DeviceID      string    `json:"device_id"`
	Name          string    `json:"name"`
	Manufacturer  string    `json:"manufacturer"`
	Model         string    `json:"model"`
	Firmware      string    `json:"firmware"`
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	RegisterTime  time.Time `json:"register_time"`
	KeepaliveTime time.Time `json:"keepalive_time"`
	Status        string    `json:"status"`
	Channels      []Channel `json:"channels"`
}

type Channel struct {
	DeviceID     string `json:"device_id"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Owner        string `json:"owner"`
	CivilCode    string `json:"civil_code"`
	Address      string `json:"address"`
	Parental     int    `json:"parental"`
	SafetyWay    int    `json:"safety_way"`
	RegisterWay  int    `json:"register_way"`
	Secrecy      int    `json:"secrecy"`
	Status       string `json:"status"`
	Longitude    string `json:"longitude"`
	Latitude     string `json:"latitude"`
	PTZType      int    `json:"ptz_type"`
}

type Stream struct {
	StreamID   string    `json:"stream_id"`
	DeviceID   string    `json:"device_id"`
	ChannelID  string    `json:"channel_id"`
	SSRC       string    `json:"ssrc"`
	RTPPort    int       `json:"rtp_port"`
	RTSPUrl    string    `json:"rtsp_url"`
	FLVUrl     string    `json:"flv_url"`
	WebRTCUrl  string    `json:"webrtc_url"`
	HLSUrl     string    `json:"hls_url"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
}

type PTZCommand struct {
	DeviceID  string `json:"device_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
	Speed     int    `json:"speed"`
}

const (
	DeviceOnline  = "ONLINE"
	DeviceOffline = "OFFLINE"
)

const (
	PTZStop      = "stop"
	PTZUp        = "up"
	PTZDown      = "down"
	PTZLeft      = "left"
	PTZRight     = "right"
	PTZZoomIn    = "zoom_in"
	PTZZoomOut   = "zoom_out"
	PTZFocusNear = "focus_near"
	PTZFocusFar  = "focus_far"
)

type InviteRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
}

type DeviceListResponse struct {
	Total   int      `json:"total"`
	Devices []Device `json:"devices"`
}

type ChannelListResponse struct {
	Total    int       `json:"total"`
	Channels []Channel `json:"channels"`
}
