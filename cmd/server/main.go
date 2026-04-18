package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wiserstream/gb28181-server/internal/api/http"
	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/services/cascade"
	"github.com/wiserstream/gb28181-server/internal/services/device"
	"github.com/wiserstream/gb28181-server/internal/services/record"
	"github.com/wiserstream/gb28181-server/internal/services/sip"
	"github.com/wiserstream/gb28181-server/internal/services/stream"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("======================================")
	fmt.Println("  GB28181 SIP Server")
	fmt.Println("======================================")
	fmt.Printf("Server ID: %s\n", cfg.Server.ID)
	fmt.Printf("Domain: %s\n", cfg.Server.Domain)
	fmt.Printf("SIP Port: %d\n", cfg.Server.SIPPort)
	fmt.Printf("HTTP Port: %d\n", cfg.Server.HTTPPort)
	fmt.Println("======================================")

	deviceSvc := device.NewDeviceService(cfg)
	streamSvc := stream.NewStreamService(cfg)
	recordSvc := record.NewRecordService()
	sipServer := sip.NewSIPServer(cfg, deviceSvc, streamSvc, recordSvc)
	cascadeSvc := cascade.NewCascadeService(cfg, sipServer)

	if err := sipServer.Start(); err != nil {
		fmt.Printf("[ERROR] Failed to start SIP server: %v\n", err)
		os.Exit(1)
	}

	httpServer := http.NewHTTPServer(cfg, deviceSvc, streamSvc, recordSvc, cascadeSvc, sipServer)

	go func() {
		if err := httpServer.Start(); err != nil {
			fmt.Printf("[ERROR] Failed to start HTTP server: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Println("[INFO] Server started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("[INFO] Shutting down server...")
	sipServer.Stop()
	httpServer.Stop()
	fmt.Println("[INFO] Server stopped")
}
