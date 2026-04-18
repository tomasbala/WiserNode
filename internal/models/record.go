package models

import "time"

type RecordItem struct {
	DeviceID   string `json:"device_id"`
	Name       string `json:"name"`
	FilePath   string `json:"file_path"`
	Address    string `json:"address"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Secrecy    int    `json:"secrecy"`
	Type       string `json:"type"`
	RecorderID string `json:"recorder_id"`
}

type RecordQuery struct {
	DeviceID  string `json:"device_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type PlayBackRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	ChannelID string `json:"channel_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

type RecordInfo struct {
	SN       int          `json:"sn"`
	DeviceID string       `json:"device_id"`
	SumNum   int          `json:"sum_num"`
	Items    []RecordItem `json:"items"`
}

type PlayBack struct {
	StreamID   string    `json:"stream_id"`
	DeviceID   string    `json:"device_id"`
	ChannelID  string    `json:"channel_id"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	SSRC       string    `json:"ssrc"`
	RTPPort    int       `json:"rtp_port"`
	RTSPUrl    string    `json:"rtsp_url"`
	FLVUrl     string    `json:"flv_url"`
	Status     string    `json:"status"`
	CreateTime time.Time `json:"create_time"`
}
