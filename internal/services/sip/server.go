package sip

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/models"
	"github.com/wiserstream/gb28181-server/internal/services/device"
	"github.com/wiserstream/gb28181-server/internal/services/record"
	"github.com/wiserstream/gb28181-server/internal/services/stream"
)

type SIPServer struct {
	cfg         *config.Config
	deviceSvc   *device.DeviceService
	streamSvc   *stream.StreamService
	recordSvc   *record.RecordService
	msgBuilder  *MessageBuilder
	udpConn     *net.UDPConn
	ctx         context.Context
	cancel      context.CancelFunc
	nonceMap    map[string]string
	callSession map[string]*CallSession
	mu          sync.RWMutex
}

type CallSession struct {
	DeviceID   string
	ChannelID  string
	CallID     string
	FromTag    string
	ToTag      string
	SSRC       string
	StreamID   string
	CreateTime time.Time
	IsPlayback bool
}

func NewSIPServer(
	cfg *config.Config,
	deviceSvc *device.DeviceService,
	streamSvc *stream.StreamService,
	recordSvc *record.RecordService,
) *SIPServer {
	return &SIPServer{
		cfg:         cfg,
		deviceSvc:   deviceSvc,
		streamSvc:   streamSvc,
		recordSvc:   recordSvc,
		msgBuilder:  NewMessageBuilder(cfg),
		nonceMap:    make(map[string]string),
		callSession: make(map[string]*CallSession),
	}
}

func (s *SIPServer) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	addr, err := net.ResolveUDPAddr("udp", s.cfg.SIPAddr())
	if err != nil {
		return fmt.Errorf("resolve udp addr failed: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen udp failed: %w", err)
	}

	s.udpConn = conn
	fmt.Printf("[SIP] Server started on %s\n", s.cfg.SIPAddr())

	go s.receiveLoop()

	return nil
}

func (s *SIPServer) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.udpConn != nil {
		s.udpConn.Close()
	}
}

func (s *SIPServer) receiveLoop() {
	buf := make([]byte, 65535)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			n, remoteAddr, err := s.udpConn.ReadFromUDP(buf)
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Printf("[SIP] Read error: %v\n", err)
				}
				continue
			}

			data := make([]byte, n)
			copy(data, buf[:n])
			go s.handleMessage(data, remoteAddr)
		}
	}
}

func (s *SIPServer) handleMessage(data []byte, remoteAddr *net.UDPAddr) {
	msgStr := string(data)
	msg := models.ParseSIPMessage(msgStr)

	deviceID := models.ExtractDeviceID(msg.From)

	fmt.Printf("[SIP] %s from %s (device: %s)\n", msg.Method, remoteAddr.String(), deviceID)

	switch msg.Method {
	case "REGISTER":
		s.handleRegister(msg, msgStr, remoteAddr)
	case "MESSAGE":
		s.handleMessageCmd(msg, msgStr, remoteAddr)
	case "INVITE":
		s.handleInvite(msg, msgStr, remoteAddr)
	case "BYE":
		s.handleBye(msg, msgStr, remoteAddr)
	case "ACK":
		s.handleAck(msg, msgStr, remoteAddr)
	default:
		if msg.StatusCode > 0 {
			s.handleResponse(msg, msgStr, remoteAddr)
		}
	}
}

func (s *SIPServer) handleRegister(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
	deviceID := models.ExtractDeviceID(msg.From)

	if msg.AuthHeader != "" {
		auth := ParseAuthHeader(msg.AuthHeader)

		s.mu.RLock()
		nonce, exists := s.nonceMap[deviceID]
		s.mu.RUnlock()

		if !exists || auth.Nonce != nonce {
			s.sendUnauthorized(msgStr, remoteAddr)
			return
		}

		if VerifyDigestAuth(auth, s.cfg.Server.Password, "REGISTER") {
			s.sendOK(msgStr, remoteAddr)

			expires := msg.Expires
			if expires <= 0 {
				s.deviceSvc.UnregisterDevice(deviceID)
				fmt.Printf("[SIP] Device %s unregistered\n", deviceID)
			} else {
				s.deviceSvc.RegisterDevice(deviceID, remoteAddr.IP.String(), remoteAddr.Port)
				fmt.Printf("[SIP] Device %s registered (expires: %d)\n", deviceID, expires)
			}
		} else {
			s.sendUnauthorized(msgStr, remoteAddr)
			fmt.Printf("[SIP] Device %s auth failed\n", deviceID)
		}
	} else {
		nonce := GenerateNonce()
		s.mu.Lock()
		s.nonceMap[deviceID] = nonce
		s.mu.Unlock()

		s.sendUnauthorized(msgStr, remoteAddr)
	}
}

func (s *SIPServer) handleMessageCmd(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
	deviceID := models.ExtractDeviceID(msg.From)

	s.sendOK(msgStr, remoteAddr)

	if msg.ContentType == "" || !strings.Contains(msg.ContentType, "MANSCDP") {
		return
	}

	cmdType := models.ParseXMLValue(msg.Content, "CmdType")

	switch cmdType {
	case "Keepalive":
		s.deviceSvc.UpdateKeepalive(deviceID)
		fmt.Printf("[SIP] Keepalive from %s\n", deviceID)

	case "Catalog":
		s.handleCatalogResponse(deviceID, msg.Content)

	case "RecordInfo":
		s.handleRecordInfoResponse(deviceID, msg.Content)

	case "DeviceStatus":
		fmt.Printf("[SIP] DeviceStatus from %s\n", deviceID)
	}
}

func (s *SIPServer) handleCatalogResponse(deviceID, content string) {
	sn := models.ParseXMLValue(content, "SN")
	sumNumStr := models.ParseXMLValue(content, "SumNum")
	sumNum, _ := strconv.Atoi(sumNumStr)

	fmt.Printf("[SIP] Catalog response from %s, SN=%s, SumNum=%d\n", deviceID, sn, sumNum)

	var channels []models.Channel

	itemStart := strings.Index(content, "<Item>")
	for itemStart != -1 {
		itemEnd := strings.Index(content, "</Item>")
		if itemEnd == -1 {
			break
		}

		item := content[itemStart : itemEnd+7]

		channel := models.Channel{
			DeviceID:     models.ParseXMLValue(item, "DeviceID"),
			Name:         models.ParseXMLValue(item, "Name"),
			Manufacturer: models.ParseXMLValue(item, "Manufacturer"),
			Model:        models.ParseXMLValue(item, "Model"),
			Owner:        models.ParseXMLValue(item, "Owner"),
			CivilCode:    models.ParseXMLValue(item, "CivilCode"),
			Address:      models.ParseXMLValue(item, "Address"),
			Status:       models.ParseXMLValue(item, "Status"),
			Longitude:    models.ParseXMLValue(item, "Longitude"),
			Latitude:     models.ParseXMLValue(item, "Latitude"),
		}

		channel.Parental, _ = strconv.Atoi(models.ParseXMLValue(item, "Parental"))
		channel.Secrecy, _ = strconv.Atoi(models.ParseXMLValue(item, "Secrecy"))

		channels = append(channels, channel)
		content = content[itemEnd+7:]
		itemStart = strings.Index(content, "<Item>")
	}

	if len(channels) > 0 {
		s.deviceSvc.UpdateChannels(deviceID, channels)
	}
}

func (s *SIPServer) handleRecordInfoResponse(deviceID, content string) {
	sn := models.ParseXMLValue(content, "SN")
	deviceIDInResp := models.ParseXMLValue(content, "DeviceID")
	sumNumStr := models.ParseXMLValue(content, "SumNum")
	sumNum, _ := strconv.Atoi(sumNumStr)

	fmt.Printf("[SIP] RecordInfo response from %s, DeviceID=%s, SN=%s, SumNum=%d\n",
		deviceID, deviceIDInResp, sn, sumNum)

	var records []models.RecordItem

	itemStart := strings.Index(content, "<Item>")
	for itemStart != -1 {
		itemEnd := strings.Index(content, "</Item>")
		if itemEnd == -1 {
			break
		}

		item := content[itemStart : itemEnd+7]

		record := models.RecordItem{
			DeviceID:   models.ParseXMLValue(item, "DeviceID"),
			Name:       models.ParseXMLValue(item, "Name"),
			FilePath:   models.ParseXMLValue(item, "FilePath"),
			Address:    models.ParseXMLValue(item, "Address"),
			StartTime:  models.ParseXMLValue(item, "StartTime"),
			EndTime:    models.ParseXMLValue(item, "EndTime"),
			RecorderID: models.ParseXMLValue(item, "RecorderID"),
		}

		record.Secrecy, _ = strconv.Atoi(models.ParseXMLValue(item, "Secrecy"))
		record.Type = models.ParseXMLValue(item, "Type")

		records = append(records, record)
		content = content[itemEnd+7:]
		itemStart = strings.Index(content, "<Item>")
	}

	if len(records) > 0 && s.recordSvc != nil {
		s.recordSvc.UpdateRecords(deviceIDInResp, records)
	}
}

func (s *SIPServer) handleInvite(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
	s.sendOK(msgStr, remoteAddr)
}

func (s *SIPServer) handleBye(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
	callID := msg.CallID

	s.sendOK(msgStr, remoteAddr)

	s.mu.Lock()
	if session, exists := s.callSession[callID]; exists {
		s.streamSvc.CloseStream(session.StreamID)
		delete(s.callSession, callID)
		fmt.Printf("[SIP] Session ended: %s\n", callID)
	}
	s.mu.Unlock()
}

func (s *SIPServer) handleAck(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
}

func (s *SIPServer) handleResponse(msg *models.SIPMessage, msgStr string, remoteAddr *net.UDPAddr) {
	fmt.Printf("[SIP] Response %d from %s\n", msg.StatusCode, remoteAddr.String())
}

func (s *SIPServer) sendOK(originalMsg string, remoteAddr *net.UDPAddr) {
	response := s.msgBuilder.BuildOK(originalMsg, models.GenerateTag())
	s.udpConn.WriteToUDP([]byte(response), remoteAddr)
}

func (s *SIPServer) sendUnauthorized(originalMsg string, remoteAddr *net.UDPAddr) {
	nonce := GenerateNonce()
	response := s.msgBuilder.BuildUnauthorized(originalMsg, nonce)
	s.udpConn.WriteToUDP([]byte(response), remoteAddr)
}

func (s *SIPServer) SendTo(data string, ip string, port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}

	_, err = s.udpConn.WriteToUDP([]byte(data), addr)
	return err
}

func (s *SIPServer) SendCatalogQuery(deviceID string) error {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	msg := s.msgBuilder.BuildCatalogQuery(deviceID)
	return s.SendTo(msg, device.IP, device.Port)
}

func (s *SIPServer) SendRecordQuery(deviceID, channelID, startTime, endTime string) error {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	sn := models.GenerateSN()
	msg := s.msgBuilder.BuildRecordQuery(deviceID, channelID, startTime, endTime, sn)
	return s.SendTo(msg, device.IP, device.Port)
}

func (s *SIPServer) SendInvite(deviceID, channelID, ssrc string, rtpPort int) (*CallSession, error) {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	callID := models.GenerateCallID(s.cfg.Server.Host)
	fromTag := models.GenerateTag()

	msg := s.msgBuilder.BuildInvite(channelID, ssrc, rtpPort, callID, fromTag)

	streamID := fmt.Sprintf("%s_%s", deviceID, channelID)

	session := &CallSession{
		DeviceID:   deviceID,
		ChannelID:  channelID,
		CallID:     callID,
		FromTag:    fromTag,
		SSRC:       ssrc,
		StreamID:   streamID,
		CreateTime: time.Now(),
		IsPlayback: false,
	}

	s.mu.Lock()
	s.callSession[callID] = session
	s.mu.Unlock()

	if err := s.SendTo(msg, device.IP, device.Port); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SIPServer) SendInvitePlayback(deviceID, channelID, ssrc string, rtpPort int, startTime, endTime string) (*CallSession, error) {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	callID := models.GenerateCallID(s.cfg.Server.Host)
	fromTag := models.GenerateTag()

	sdp := s.msgBuilder.BuildSDPPlayback(ssrc, rtpPort, startTime, endTime)

	cseq := time.Now().Unix()
	branch := models.GenerateBranch()

	msg := fmt.Sprintf(
		"INVITE sip:%s@%s SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP %s:%d;rport;branch=%s\r\n"+
			"From: <sip:%s@%s>;tag=%s\r\n"+
			"To: <sip:%s@%s>\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %d INVITE\r\n"+
			"Contact: <sip:%s@%s:%d>\r\n"+
			"Max-Forwards: 70\r\n"+
			"User-Agent: GB28181-Server\r\n"+
			"Content-Type: application/sdp\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		channelID, s.cfg.Server.Domain,
		s.cfg.Server.Host, s.cfg.Server.SIPPort, branch,
		s.cfg.Server.ID, s.cfg.Server.Domain, fromTag,
		channelID, s.cfg.Server.Domain,
		callID,
		cseq,
		s.cfg.Server.ID, s.cfg.Server.Host, s.cfg.Server.SIPPort,
		len(sdp), sdp,
	)

	streamID := fmt.Sprintf("playback_%s_%s", deviceID, channelID)

	session := &CallSession{
		DeviceID:   deviceID,
		ChannelID:  channelID,
		CallID:     callID,
		FromTag:    fromTag,
		SSRC:       ssrc,
		StreamID:   streamID,
		CreateTime: time.Now(),
		IsPlayback: true,
	}

	s.mu.Lock()
	s.callSession[callID] = session
	s.mu.Unlock()

	if err := s.SendTo(msg, device.IP, device.Port); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SIPServer) SendBye(deviceID, channelID string) error {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for callID, session := range s.callSession {
		if session.DeviceID == deviceID && session.ChannelID == channelID {
			msg := s.msgBuilder.BuildBye(channelID, callID, session.FromTag, session.ToTag)
			delete(s.callSession, callID)
			return s.SendTo(msg, device.IP, device.Port)
		}
	}

	return nil
}

func (s *SIPServer) SendPTZControl(deviceID, channelID, action string, speed int) error {
	device, exists := s.deviceSvc.GetDevice(deviceID)
	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	ptzCmd := s.msgBuilder.GeneratePTZCommand(action, speed)
	msg := s.msgBuilder.BuildPTZControl(deviceID, channelID, ptzCmd)

	return s.SendTo(msg, device.IP, device.Port)
}

func (s *SIPServer) GetSession(callID string) (*CallSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.callSession[callID]
	return session, exists
}

func (s *SIPServer) UpdateSessionToTag(callID, toTag string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.callSession[callID]; exists {
		session.ToTag = toTag
	}
}
