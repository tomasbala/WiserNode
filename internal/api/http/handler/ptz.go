package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wiserstream/gb28181-server/internal/models"
	"github.com/wiserstream/gb28181-server/internal/services/sip"
)

type PTZHandler struct {
	sipServer *sip.SIPServer
}

func NewPTZHandler(sipServer *sip.SIPServer) *PTZHandler {
	return &PTZHandler{sipServer: sipServer}
}

func (h *PTZHandler) Control(c *gin.Context) {
	var req models.PTZCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	speed := req.Speed
	if speed <= 0 {
		speed = 5
	}

	if err := h.sipServer.SendPTZControl(req.DeviceID, req.ChannelID, req.Action, speed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
