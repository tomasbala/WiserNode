package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "video-platform/internal/api/http"
	"video-platform/internal/protocols/onvif"
	"video-platform/internal/protocols/gb28181"
	"video-platform/internal/protocols/rtsp"
	"video-platform/internal/services/device"
	"video-platform/internal/services/stream"
	"video-platform/internal/services/session"

	"github.com/gin-gonic/gin"
)

var (
	grpcPort = flag.Int("grpc-port", 50051, "gRPC server port")
	httpPort = flag.Int("http-port", 8080, "HTTP server port")
	debug    = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	// 设置Gin模式
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化服务
	deviceService := device.NewService()
	streamService := stream.NewService()
	sessionService := session.NewService()

	// 初始化协议模块
	onvifProtocol := onvif.NewProtocol(deviceService)
	gb28181Protocol := gb28181.NewProtocol(deviceService)
	rtspProtocol := rtsp.NewProtocol(streamService)

	// 启动协议模块
	go onvifProtocol.Start()
	go gb28181Protocol.Start()
	go rtspProtocol.Start()

	// 初始化HTTP服务
	httpServer := apihttp.NewServer(
		deviceService,
		streamService,
		sessionService,
	)

	// 启动HTTP服务
	httpAddr := fmt.Sprintf(":%d", *httpPort)
	httpServerAddr := &http.Server{
		Addr:    httpAddr,
		Handler: httpServer.Router(),
	}

	go func() {
		log.Printf("HTTP server started on %s", httpAddr)
		if err := httpServerAddr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务
	if err := httpServerAddr.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server forced to shutdown: %v", err)
	}

	// 停止协议模块
	onvifProtocol.Stop()
	gb28181Protocol.Stop()
	rtspProtocol.Stop()

	log.Println("Servers exited properly")
}
