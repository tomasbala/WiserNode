package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wiserstream/gb28181-server/internal/api/http/handler"
	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/services/cascade"
	"github.com/wiserstream/gb28181-server/internal/services/device"
	"github.com/wiserstream/gb28181-server/internal/services/record"
	"github.com/wiserstream/gb28181-server/internal/services/sip"
	"github.com/wiserstream/gb28181-server/internal/services/stream"
)

type HTTPServer struct {
	cfg        *config.Config
	deviceSvc  *device.DeviceService
	streamSvc  *stream.StreamService
	recordSvc  *record.RecordService
	cascadeSvc *cascade.CascadeService
	sipServer  *sip.SIPServer
	engine     *gin.Engine
}

func NewHTTPServer(
	cfg *config.Config,
	deviceSvc *device.DeviceService,
	streamSvc *stream.StreamService,
	recordSvc *record.RecordService,
	cascadeSvc *cascade.CascadeService,
	sipServer *sip.SIPServer,
) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)

	return &HTTPServer{
		cfg:        cfg,
		deviceSvc:  deviceSvc,
		streamSvc:  streamSvc,
		recordSvc:  recordSvc,
		cascadeSvc: cascadeSvc,
		sipServer:  sipServer,
		engine:     gin.New(),
	}
}

func (s *HTTPServer) SetupRoutes() {
	deviceHandler := handler.NewDeviceHandler(s.deviceSvc)
	streamHandler := handler.NewStreamHandler(s.streamSvc, s.recordSvc, s.sipServer)
	ptzHandler := handler.NewPTZHandler(s.sipServer)
	cascadeHandler := handler.NewCascadeHandler(s.cascadeSvc)

	api := s.engine.Group("/api/v1")
	{
		api.GET("/devices", deviceHandler.ListDevices)
		api.GET("/devices/:device_id", deviceHandler.GetDevice)
		api.GET("/devices/:device_id/channels", deviceHandler.GetChannels)
		api.POST("/devices/:device_id/catalog", streamHandler.QueryCatalog)

		api.POST("/stream/play", streamHandler.StartPlay)
		api.DELETE("/stream/play/:device_id/:channel_id", streamHandler.StopPlay)
		api.GET("/streams", streamHandler.ListStreams)
		api.GET("/streams/:stream_id", streamHandler.GetStream)

		api.POST("/stream/playback", streamHandler.StartPlayback)
		api.DELETE("/stream/playback/:device_id/:channel_id", streamHandler.StopPlayback)

		api.POST("/records/query", streamHandler.QueryRecords)
		api.GET("/records/:channel_id", streamHandler.GetRecords)

		api.POST("/ptz/control", ptzHandler.Control)

		api.GET("/cascade/platforms", cascadeHandler.ListPlatforms)
		api.POST("/cascade/platforms", cascadeHandler.AddPlatform)
		api.DELETE("/cascade/platforms/:platform_id", cascadeHandler.RemovePlatform)
		api.POST("/cascade/platforms/:platform_id/register", cascadeHandler.RegisterPlatform)
		api.POST("/cascade/platforms/:platform_id/catalog", cascadeHandler.PushCatalog)
	}

	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
}

func (s *HTTPServer) Start() error {
	s.SetupRoutes()
	return s.engine.Run(s.cfg.HTTPAddr())
}

func (s *HTTPServer) Stop() {
}
