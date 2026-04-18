package cascade

import (
	"fmt"
	"sync"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/models"
	"github.com/wiserstream/gb28181-server/internal/services/sip"
)

type CascadeService struct {
	cfg       *config.Config
	sipServer *sip.SIPServer
	platforms map[string]*models.CascadePlatform
	channels  map[string][]models.CascadeChannel
	mu        sync.RWMutex
}

func NewCascadeService(cfg *config.Config, sipServer *sip.SIPServer) *CascadeService {
	return &CascadeService{
		cfg:       cfg,
		sipServer: sipServer,
		platforms: make(map[string]*models.CascadePlatform),
		channels:  make(map[string][]models.CascadeChannel),
	}
}

func (s *CascadeService) AddPlatform(platform *models.CascadePlatform) {
	s.mu.Lock()
	defer s.mu.Unlock()

	platform.Status = models.CascadeOffline
	s.platforms[platform.ID] = platform
}

func (s *CascadeService) RemovePlatform(platformID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.platforms, platformID)
	delete(s.channels, platformID)
}

func (s *CascadeService) GetPlatform(platformID string) (*models.CascadePlatform, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	platform, exists := s.platforms[platformID]
	return platform, exists
}

func (s *CascadeService) ListPlatforms() []*models.CascadePlatform {
	s.mu.RLock()
	defer s.mu.RUnlock()

	platforms := make([]*models.CascadePlatform, 0, len(s.platforms))
	for _, p := range s.platforms {
		platforms = append(platforms, p)
	}
	return platforms
}

func (s *CascadeService) UpdatePlatformStatus(platformID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if platform, exists := s.platforms[platformID]; exists {
		platform.Status = status
		if status == models.CascadeOnline {
			platform.RegisterTime = time.Now()
		}
		platform.KeepaliveTime = time.Now()
	}
}

func (s *CascadeService) RegisterToPlatform(platformID string) error {
	s.mu.RLock()
	platform, exists := s.platforms[platformID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("platform %s not found", platformID)
	}

	msgBuilder := sip.NewMessageBuilder(s.cfg)
	registerMsg := msgBuilder.BuildRegister(platform.ServerID, platform.Expires, "")

	return s.sipServer.SendTo(registerMsg, platform.ServerIP, platform.ServerPort)
}

func (s *CascadeService) SendKeepalive(platformID string) error {
	s.mu.RLock()
	platform, exists := s.platforms[platformID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("platform %s not found", platformID)
	}

	msgBuilder := sip.NewMessageBuilder(s.cfg)
	keepaliveMsg := msgBuilder.BuildNotifyKeepalive(platform.ServerID)

	return s.sipServer.SendTo(keepaliveMsg, platform.ServerIP, platform.ServerPort)
}

func (s *CascadeService) PushCatalog(platformID string, channels []models.Channel) error {
	s.mu.RLock()
	platform, exists := s.platforms[platformID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("platform %s not found", platformID)
	}

	sn := models.GenerateSN()

	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Response>
<CmdType>Catalog</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<SumNum>%d</SumNum>
`,
		sn, s.cfg.Server.ID, len(channels),
	)

	for _, ch := range channels {
		body += fmt.Sprintf(
			`<Item>
<DeviceID>%s</DeviceID>
<Name>%s</Name>
<Manufacturer>%s</Manufacturer>
<Model>%s</Model>
<Owner>%s</Owner>
<CivilCode>%s</CivilCode>
<Address>%s</Address>
<Parental>0</Parental>
<SafetyWay>0</SafetyWay>
<RegisterWay>1</RegisterWay>
<Secrecy>0</Secrecy>
<Status>%s</Status>
<Longitude>%s</Longitude>
<Latitude>%s</Latitude>
</Item>
`,
			ch.DeviceID, ch.Name, ch.Manufacturer, ch.Model, ch.Owner,
			ch.CivilCode, ch.Address, ch.Status, ch.Longitude, ch.Latitude,
		)
	}

	body += "</Response>"

	msgBuilder := sip.NewMessageBuilder(s.cfg)
	catalogMsg := msgBuilder.BuildMessage(platform.ServerID, body, "Application/MANSCDP+xml")

	return s.sipServer.SendTo(catalogMsg, platform.ServerIP, platform.ServerPort)
}

func (s *CascadeService) AddChannel(platformID string, channel *models.CascadeChannel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel.PlatformID = platformID
	s.channels[platformID] = append(s.channels[platformID], *channel)
}

func (s *CascadeService) GetChannels(platformID string) []models.CascadeChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels, exists := s.channels[platformID]
	if !exists {
		return nil
	}

	result := make([]models.CascadeChannel, len(channels))
	copy(result, channels)
	return result
}

func (s *CascadeService) RemoveChannel(platformID, channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channels, exists := s.channels[platformID]
	if !exists {
		return
	}

	for i, ch := range channels {
		if ch.ChannelID == channelID {
			s.channels[platformID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
}

func (s *CascadeService) StartHeartbeat(platformID string) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			s.mu.RLock()
			_, exists := s.platforms[platformID]
			s.mu.RUnlock()

			if !exists {
				return
			}

			if err := s.SendKeepalive(platformID); err != nil {
				fmt.Printf("[Cascade] Keepalive failed for %s: %v\n", platformID, err)
			}
		}
	}()
}
