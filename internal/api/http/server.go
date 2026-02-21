package http

import (
	"net/http"
	"strconv"

	"video-platform/internal/models"
	"video-platform/internal/services/device"
	"video-platform/internal/services/stream"
	"video-platform/internal/services/session"

	"github.com/gin-gonic/gin"
)

// Server HTTP服务实现
type Server struct {
	deviceService  device.Service
	streamService  stream.Service
	sessionService session.Service
}

// NewServer 创建HTTP服务实例
func NewServer(
	deviceService device.Service,
	streamService stream.Service,
	sessionService session.Service,
) *Server {
	return &Server{
		deviceService:  deviceService,
		streamService:  streamService,
		sessionService: sessionService,
	}
}

// Router 路由配置
func (s *Server) Router() *gin.Engine {
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务
	r.StaticFile("/", "web/index.html")
	r.Static("/web", "web")

	// API路由组
	api := r.Group("/api")
	{
		// 设备管理
		devices := api.Group("/devices")
		{
			devices.GET("", s.GetDevices)
			devices.POST("", s.AddDevice)
			devices.GET("/:id", s.GetDevice)
			devices.DELETE("/:id", s.DeleteDevice)
			devices.GET("/:id/status", s.GetDeviceStatus)
			devices.GET("/:id/discover", s.DiscoverDevices)
		}

		// 流媒体管理
		streams := api.Group("/streams")
		{
			streams.GET("", s.GetStreams)
			streams.POST("", s.StartStream)
			streams.GET("/:id", s.GetStream)
			streams.DELETE("/:id", s.StopStream)
			streams.GET("/:id/url", s.GetStreamURL)
		}

		// 会话管理
		sessions := api.Group("/sessions")
		{
			sessions.GET("", s.GetSessions)
			sessions.POST("", s.CreateSession)
			sessions.GET("/:id", s.GetSession)
			sessions.DELETE("/:id", s.CloseSession)
		}
	}

	return r
}

// GetDevices 获取设备列表
func (s *Server) GetDevices(c *gin.Context) {
	devices, err := s.deviceService.GetDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// AddDevice 添加设备
func (s *Server) AddDevice(c *gin.Context) {
	var req models.AddDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := s.deviceService.AddDevice(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

// GetDevice 获取设备
func (s *Server) GetDevice(c *gin.Context) {
	id := c.Param("id")
	device, err := s.deviceService.GetDevice(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// DeleteDevice 删除设备
func (s *Server) DeleteDevice(c *gin.Context) {
	id := c.Param("id")
	err := s.deviceService.DeleteDevice(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetDeviceStatus 获取设备状态
func (s *Server) GetDeviceStatus(c *gin.Context) {
	id := c.Param("id")
	status, err := s.deviceService.GetDeviceStatus(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// DiscoverDevices 发现设备
func (s *Server) DiscoverDevices(c *gin.Context) {
	timeoutStr := c.DefaultQuery("timeout", "5")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 5
	}

	devices, err := s.deviceService.DiscoverDevices(timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// GetStreams 获取流列表
func (s *Server) GetStreams(c *gin.Context) {
	streams, err := s.streamService.GetStreams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, streams)
}

// StartStream 启动流
func (s *Server) StartStream(c *gin.Context) {
	var req models.StartStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := s.streamService.StartStream(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, stream)
}

// GetStream 获取流
func (s *Server) GetStream(c *gin.Context) {
	id := c.Param("id")
	stream, err := s.streamService.GetStream(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// StopStream 停止流
func (s *Server) StopStream(c *gin.Context) {
	id := c.Param("id")
	err := s.streamService.StopStream(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetStreamURL 获取流地址
func (s *Server) GetStreamURL(c *gin.Context) {
	id := c.Param("id")
	url, err := s.streamService.GetStreamURL(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetSessions 获取会话列表
func (s *Server) GetSessions(c *gin.Context) {
	sessions, err := s.sessionService.GetSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// CreateSession 创建会话
func (s *Server) CreateSession(c *gin.Context) {
	var req struct {
		StreamID string `json:"stream_id" binding:"required"`
		ClientID string `json:"client_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := s.sessionService.CreateSession(req.StreamID, req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetSession 获取会话
func (s *Server) GetSession(c *gin.Context) {
	id := c.Param("id")
	session, err := s.sessionService.GetSession(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// CloseSession 关闭会话
func (s *Server) CloseSession(c *gin.Context) {
	id := c.Param("id")
	err := s.sessionService.CloseSession(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
