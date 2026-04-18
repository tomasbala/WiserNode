package models

import "time"

type CascadePlatform struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ServerID      string    `json:"server_id"`
	ServerDomain  string    `json:"server_domain"`
	ServerIP      string    `json:"server_ip"`
	ServerPort    int       `json:"server_port"`
	Username      string    `json:"username"`
	Password      string    `json:"password"`
	Expires       int       `json:"expires"`
	Status        string    `json:"status"`
	RegisterTime  time.Time `json:"register_time"`
	KeepaliveTime time.Time `json:"keepalive_time"`
}

type CascadeChannel struct {
	ID         string `json:"id"`
	PlatformID string `json:"platform_id"`
	ChannelID  string `json:"channel_id"`
	DeviceID   string `json:"device_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	StreamID   string `json:"stream_id"`
}

type CascadeRegister struct {
	PlatformID string `json:"platform_id" binding:"required"`
}

const (
	CascadeOnline  = "ONLINE"
	CascadeOffline = "OFFLINE"
)

type CatalogPushRequest struct {
	PlatformID string `json:"platform_id" binding:"required"`
}
