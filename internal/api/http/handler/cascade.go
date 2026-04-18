package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wiserstream/gb28181-server/internal/models"
	"github.com/wiserstream/gb28181-server/internal/services/cascade"
)

type CascadeHandler struct {
	cascadeSvc *cascade.CascadeService
}

func NewCascadeHandler(cascadeSvc *cascade.CascadeService) *CascadeHandler {
	return &CascadeHandler{cascadeSvc: cascadeSvc}
}

func (h *CascadeHandler) ListPlatforms(c *gin.Context) {
	platforms := h.cascadeSvc.ListPlatforms()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     len(platforms),
			"platforms": platforms,
		},
	})
}

func (h *CascadeHandler) AddPlatform(c *gin.Context) {
	var platform models.CascadePlatform
	if err := c.ShouldBindJSON(&platform); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	h.cascadeSvc.AddPlatform(&platform)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "platform added",
	})
}

func (h *CascadeHandler) RemovePlatform(c *gin.Context) {
	platformID := c.Param("platform_id")

	h.cascadeSvc.RemovePlatform(platformID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "platform removed",
	})
}

func (h *CascadeHandler) RegisterPlatform(c *gin.Context) {
	platformID := c.Param("platform_id")

	if err := h.cascadeSvc.RegisterToPlatform(platformID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	h.cascadeSvc.UpdatePlatformStatus(platformID, models.CascadeOnline)
	h.cascadeSvc.StartHeartbeat(platformID)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "register sent",
	})
}

func (h *CascadeHandler) PushCatalog(c *gin.Context) {
	platformID := c.Param("platform_id")

	var req struct {
		Channels []models.Channel `json:"channels" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request",
		})
		return
	}

	if err := h.cascadeSvc.PushCatalog(platformID, req.Channels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "catalog pushed",
	})
}
