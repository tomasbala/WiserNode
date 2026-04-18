package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wiserstream/gb28181-server/internal/models"
	"github.com/wiserstream/gb28181-server/internal/services/record"
	"github.com/wiserstream/gb28181-server/internal/services/sip"
	"github.com/wiserstream/gb28181-server/internal/services/stream"
)

type StreamHandler struct {
	streamSvc *stream.StreamService
	recordSvc *record.RecordService
	sipServer *sip.SIPServer
}

func NewStreamHandler(streamSvc *stream.StreamService, recordSvc *record.RecordService, sipServer *sip.SIPServer) *StreamHandler {
	return &StreamHandler{
		streamSvc: streamSvc,
		recordSvc: recordSvc,
		sipServer: sipServer,
	}
}

func (h *StreamHandler) StartPlay(c *gin.Context) {
	var req models.InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	ssrc := h.generateSSRC()

	st, err := h.streamSvc.OpenStream(req.DeviceID, req.ChannelID, ssrc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to open stream: %v", err),
		})
		return
	}

	_, err = h.sipServer.SendInvite(req.DeviceID, req.ChannelID, ssrc, st.RTPPort)
	if err != nil {
		h.streamSvc.CloseStream(st.StreamID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to send invite: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"stream_id":  st.StreamID,
			"ssrc":       ssrc,
			"rtp_port":   st.RTPPort,
			"rtsp_url":   st.RTSPUrl,
			"flv_url":    st.FLVUrl,
			"hls_url":    st.HLSUrl,
			"webrtc_url": st.WebRTCUrl,
		},
	})
}

func (h *StreamHandler) StopPlay(c *gin.Context) {
	deviceID := c.Param("device_id")
	channelID := c.Param("channel_id")
	streamID := fmt.Sprintf("%s_%s", deviceID, channelID)

	if err := h.sipServer.SendBye(deviceID, channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to send bye: %v", err),
		})
		return
	}

	h.streamSvc.CloseStream(streamID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *StreamHandler) GetStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	st, exists := h.streamSvc.GetStream(streamID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "stream not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": st,
	})
}

func (h *StreamHandler) ListStreams(c *gin.Context) {
	streams := h.streamSvc.ListStreams()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":   len(streams),
			"streams": streams,
		},
	})
}

func (h *StreamHandler) QueryCatalog(c *gin.Context) {
	deviceID := c.Param("device_id")

	if err := h.sipServer.SendCatalogQuery(deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to query catalog: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "catalog query sent",
	})
}

func (h *StreamHandler) QueryRecords(c *gin.Context) {
	var req models.RecordQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	if req.StartTime == "" {
		req.StartTime = time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05")
	}
	if req.EndTime == "" {
		req.EndTime = time.Now().Format("2006-01-02T15:04:05")
	}

	if err := h.sipServer.SendRecordQuery(req.DeviceID, req.ChannelID, req.StartTime, req.EndTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to query records: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "record query sent",
	})
}

func (h *StreamHandler) GetRecords(c *gin.Context) {
	channelID := c.Param("channel_id")

	records := h.recordSvc.GetRecords(channelID)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":   len(records),
			"records": records,
		},
	})
}

func (h *StreamHandler) StartPlayback(c *gin.Context) {
	var req models.PlayBackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	ssrc := h.generateSSRC()

	st, err := h.streamSvc.OpenStream("playback_"+req.DeviceID, req.ChannelID, ssrc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to open stream: %v", err),
		})
		return
	}

	_, err = h.sipServer.SendInvitePlayback(req.DeviceID, req.ChannelID, ssrc, st.RTPPort, req.StartTime, req.EndTime)
	if err != nil {
		h.streamSvc.CloseStream(st.StreamID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to start playback: %v", err),
		})
		return
	}

	pb := h.recordSvc.CreatePlayback(req.DeviceID, req.ChannelID, ssrc, st.RTPPort, req.StartTime, req.EndTime)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"stream_id":  pb.StreamID,
			"ssrc":       ssrc,
			"rtp_port":   st.RTPPort,
			"rtsp_url":   st.RTSPUrl,
			"flv_url":    st.FLVUrl,
			"hls_url":    st.HLSUrl,
			"start_time": req.StartTime,
			"end_time":   req.EndTime,
		},
	})
}

func (h *StreamHandler) StopPlayback(c *gin.Context) {
	deviceID := c.Param("device_id")
	channelID := c.Param("channel_id")

	if err := h.sipServer.SendBye(deviceID, channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": fmt.Sprintf("failed to stop playback: %v", err),
		})
		return
	}

	streamID := fmt.Sprintf("playback_%s_%s", deviceID, channelID)
	h.streamSvc.CloseStream(streamID)
	h.recordSvc.RemovePlayback(streamID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func (h *StreamHandler) generateSSRC() string {
	timestamp := time.Now().UnixNano() % 1000000000
	return fmt.Sprintf("%010d", timestamp)
}
