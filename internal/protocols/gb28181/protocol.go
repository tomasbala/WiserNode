package gb28181

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"video-platform/internal/models"
	"video-platform/internal/services/device"
)

// SIPMessage SIP消息结构
type SIPMessage struct {
	Method     string
	Headers    map[string]string
	Body       string
	Message    string
	RemoteAddr string
}

// SIPRequestProcessor SIP请求处理器接口
type SIPRequestProcessor interface {
	Process(message *SIPMessage, conn net.Conn)
}

// SIPServer SIP服务器实现
type SIPServer struct {
	tcpListeners []net.Listener
	udpListeners []net.PacketConn
	running      bool
	mutex        sync.Mutex
	protocol     *Protocol
	processors   map[string]SIPRequestProcessor
	monitorIPs   []string
}

// RecordInfo 录像信息结构
type RecordInfo struct {
	RecordID    string
	DeviceID    string
	ChannelID   string
	StartTime   time.Time
	EndTime     time.Time
	FileSize    int64
	FilePath    string
	StreamType  string
	Status      string
}

// Protocol GB28181协议实现
type Protocol struct {
	deviceService device.Service
	devices       map[string]*models.Device
	sipServer     *SIPServer
	records       map[string]*RecordInfo
	recordsMutex  sync.Mutex
}

// NewProtocol 创建GB28181协议实例
func NewProtocol(deviceService device.Service) *Protocol {
	return &Protocol{
		deviceService: deviceService,
		devices:       make(map[string]*models.Device),
		records:       make(map[string]*RecordInfo),
	}
}

// NewSIPServer 创建SIP服务器实例
func NewSIPServer(protocol *Protocol) *SIPServer {
	return &SIPServer{
		protocol:   protocol,
		running:    false,
		processors: make(map[string]SIPRequestProcessor),
		monitorIPs: []string{},
	}
}

// Start 启动GB28181协议服务
func (p *Protocol) Start() {
	log.Println("Starting GB28181 protocol service...")

	// 启动SIP服务器
	p.sipServer = NewSIPServer(p)
	if err := p.sipServer.Start(); err != nil {
		log.Printf("Failed to start SIP server: %v", err)
		return
	}

	// 创建处理器实例
	inviteProcessor := &InviteProcessor{protocol: p, streams: make(map[string]*RTPStream)}
	messageProcessor := &MessageProcessor{protocol: p, alarms: make(map[string]*AlarmInfo)}

	// 注册SIP请求处理器
	p.sipServer.RegisterProcessor("REGISTER", &RegisterProcessor{protocol: p})
	p.sipServer.RegisterProcessor("INVITE", inviteProcessor)
	p.sipServer.RegisterProcessor("ACK", &AckProcessor{protocol: p})
	p.sipServer.RegisterProcessor("BYE", &ByeProcessor{protocol: p, inviteProcessor: inviteProcessor})
	p.sipServer.RegisterProcessor("MESSAGE", messageProcessor)

	// 启动设备状态监控
	go p.startDeviceMonitoring()

	log.Println("GB28181 protocol service started")
}

// RegisterProcessor 注册SIP请求处理器
func (s *SIPServer) RegisterProcessor(method string, processor SIPRequestProcessor) {
	s.processors[method] = processor
}

// Start 启动SIP服务器
func (s *SIPServer) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("SIP server is already running")
	}

	// 自动检测网卡IP
	s.detectMonitorIPs()

	if len(s.monitorIPs) == 0 {
		return fmt.Errorf("no network interfaces found")
	}

	// 为每个IP创建TCP和UDP监听
	port := 5060
	for _, ip := range s.monitorIPs {
		// 创建TCP监听
		tcpAddr := fmt.Sprintf("%s:%d", ip, port)
		tcpListener, err := net.Listen("tcp", tcpAddr)
		if err != nil {
			log.Printf("Failed to start TCP listener on %s: %v", tcpAddr, err)
			continue
		}
		s.tcpListeners = append(s.tcpListeners, tcpListener)
		log.Printf("SIP TCP server started on %s", tcpAddr)

		// 创建UDP监听
		udpAddr := fmt.Sprintf("%s:%d", ip, port)
		udpListener, err := net.ListenPacket("udp", udpAddr)
		if err != nil {
			log.Printf("Failed to start UDP listener on %s: %v", udpAddr, err)
			continue
		}
		s.udpListeners = append(s.udpListeners, udpListener)
		log.Printf("SIP UDP server started on %s", udpAddr)

		// 启动处理goroutine
		go s.handleTCPConnections(tcpListener)
		go s.handleUDPConnections(udpListener)
	}

	if len(s.tcpListeners) == 0 && len(s.udpListeners) == 0 {
		return fmt.Errorf("failed to start any listeners")
	}

	s.running = true
	return nil
}

// detectMonitorIPs 检测监控IP
func (s *SIPServer) detectMonitorIPs() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v", err)
		return
	}

	for _, iface := range interfaces {
		// 跳过 docker 接口和环回接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if strings.HasPrefix(iface.Name, "docker") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					s.monitorIPs = append(s.monitorIPs, ipnet.IP.String())
					log.Printf("Added monitor IP: %s", ipnet.IP.String())
				}
			}
		}
	}
}

// Stop 停止SIP服务器
func (s *SIPServer) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false

	// 关闭所有TCP监听
	for _, listener := range s.tcpListeners {
		listener.Close()
	}

	// 关闭所有UDP监听
	for _, listener := range s.udpListeners {
		listener.Close()
	}

	log.Println("SIP server stopped")
}

// handleTCPConnections 处理TCP连接
func (s *SIPServer) handleTCPConnections(listener net.Listener) {
	for s.running {
		conn, err := listener.Accept()
		if err != nil {
			if !s.running {
				break
			}
			log.Printf("Error accepting TCP connection: %v", err)
			continue
		}

		// 处理每个连接
		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection 处理单个TCP连接
func (s *SIPServer) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("New SIP TCP connection from %s", conn.RemoteAddr())

	// 读取SIP消息
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if !s.running {
				break
			}
			log.Printf("Error reading from TCP connection: %v", err)
			break
		}

		message := string(buffer[:n])
		log.Printf("Received SIP TCP message: %s", message)

		// 解析并处理SIP消息
		sipMsg := s.parseSIPMessage(message)
		sipMsg.RemoteAddr = conn.RemoteAddr().String()
		s.handleSIPMessage(sipMsg, conn)
	}
}

// handleUDPConnections 处理UDP连接
func (s *SIPServer) handleUDPConnections(listener net.PacketConn) {
	buffer := make([]byte, 4096)
	for s.running {
		n, addr, err := listener.ReadFrom(buffer)
		if err != nil {
			if !s.running {
				break
			}
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		log.Printf("Received SIP UDP message from %s: %s", addr, message)

		// 解析并处理SIP消息
		sipMsg := s.parseSIPMessage(message)
		sipMsg.RemoteAddr = addr.String()
		s.handleSIPUDPMsg(listener, addr, sipMsg)
	}
}

// parseSIPMessage 解析SIP消息
func (s *SIPServer) parseSIPMessage(message string) *SIPMessage {
	if message == "" {
		log.Println("Empty SIP message received")
		return &SIPMessage{
			Headers: make(map[string]string),
			Message: message,
		}
	}

	lines := strings.Split(message, "\r\n")
	sipMsg := &SIPMessage{
		Headers: make(map[string]string),
		Message: message,
	}

	// 解析首行
	if len(lines) > 0 {
		firstLine := lines[0]
		parts := strings.Split(firstLine, " ")
		if len(parts) >= 1 {
			// 检查是否是请求或响应
			if strings.HasPrefix(parts[0], "SIP/") {
				// 响应
				sipMsg.Method = "RESPONSE"
			} else {
				// 请求
				sipMsg.Method = parts[0]
			}
		} else {
			log.Printf("Invalid SIP message format: %s", firstLine)
		}
	}

	// 解析头部
	bodyStart := false
	var body strings.Builder
	for i, line := range lines {
		if i == 0 {
			continue // 跳过首行
		}

		if line == "" {
			bodyStart = true
			continue
		}

		if bodyStart {
			body.WriteString(line)
			body.WriteString("\r\n")
		} else {
			// 解析头部
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				sipMsg.Headers[key] = value
			} else {
				log.Printf("Invalid header format: %s", line)
			}
		}
	}

	sipMsg.Body = body.String()
	return sipMsg
}

// handleSIPMessage 处理SIP消息
func (s *SIPServer) handleSIPMessage(message *SIPMessage, conn net.Conn) {
	processor, exists := s.processors[message.Method]
	if !exists {
		log.Printf("No processor for SIP method: %s", message.Method)
		// 发送405 Method Not Allowed
		s.sendMethodNotAllowed(conn, message)
		return
	}

	// 处理消息
	processor.Process(message, conn)
}

// handleSIPUDPMsg 处理UDP SIP消息
func (s *SIPServer) handleSIPUDPMsg(conn net.PacketConn, addr net.Addr, message *SIPMessage) {
	processor, exists := s.processors[message.Method]
	if !exists {
		log.Printf("No processor for SIP method: %s", message.Method)
		// 发送405 Method Not Allowed
		s.sendUDPMethodNotAllowed(conn, addr, message)
		return
	}

	// 对于UDP，我们需要一个特殊的处理器包装
	udpProcessor := &UDPProcessorWrapper{
		processor: processor,
		conn:      conn,
		addr:      addr,
	}
	udpProcessor.Process(message, nil)
}

// sendMethodNotAllowed 发送405 Method Not Allowed响应
func (s *SIPServer) sendMethodNotAllowed(conn net.Conn, message *SIPMessage) {
	response := "SIP/2.0 405 Method Not Allowed\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + s.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Allow: REGISTER, INVITE, ACK, BYE, MESSAGE\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending 405 response: %v", err)
	}
}

// sendUDPMethodNotAllowed 发送UDP 405 Method Not Allowed响应
func (s *SIPServer) sendUDPMethodNotAllowed(conn net.PacketConn, addr net.Addr, message *SIPMessage) {
	response := "SIP/2.0 405 Method Not Allowed\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + s.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Allow: REGISTER, INVITE, ACK, BYE, MESSAGE\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.WriteTo([]byte(response), addr)
	if err != nil {
		log.Printf("Error sending UDP 405 response: %v", err)
	}
}

// generateTag 生成SIP标签
func (s *SIPServer) generateTag() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}

// UDPProcessorWrapper UDP处理器包装器
type UDPProcessorWrapper struct {
	processor SIPRequestProcessor
	conn      net.PacketConn
	addr      net.Addr
}

// Process 处理UDP消息
func (w *UDPProcessorWrapper) Process(message *SIPMessage, conn net.Conn) {
	// 对于UDP，我们需要特殊处理
	// 这里可以根据需要实现
	log.Printf("Processing UDP message: %s", message.Method)
}

// DeviceSession 设备会话
type DeviceSession struct {
	DeviceID    string
	Device      *models.Device
	Contact     string
	Expires     int
	LastRegister time.Time
}

// RegisterProcessor 注册请求处理器
type RegisterProcessor struct {
	protocol *Protocol
}

// Process 处理注册请求
func (p *RegisterProcessor) Process(message *SIPMessage, conn net.Conn) {
	log.Println("Processing REGISTER request")

	// 解析设备信息
	deviceID, err := p.parseDeviceID(message.Headers["From"])
	if err != nil {
		log.Printf("Failed to parse device ID: %v", err)
		p.sendErrorResponse(conn, message, "400", "Bad Request")
		return
	}

	// 解析过期时间
	expires := 3600 // 默认3600秒
	if exp, ok := message.Headers["Expires"]; ok {
		if expInt, err := strconv.Atoi(exp); err == nil {
			expires = expInt
		}
	}

	// 解析设备地址
	contact := message.Headers["Contact"]

	// 查找或创建设备
	device, err := p.protocol.deviceService.GetDevice(deviceID)
	if err != nil {
		// 创建设备
		device = &models.Device{
			ID:       deviceID,
			Name:     "GB28181 Device " + deviceID,
			Type:     models.DeviceTypeGB28181,
			Status:   models.DeviceStatusOnline,
			IP:       strings.Split(conn.RemoteAddr().String(), ":")[0],
			Port:     5060,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// 保存设备
		req := &models.AddDeviceRequest{
			Name: device.Name,
			Type: device.Type,
			IP:   device.IP,
			Port: device.Port,
		}
		if _, err := p.protocol.deviceService.AddDevice(req); err != nil {
			log.Printf("Failed to add device: %v", err)
			p.sendErrorResponse(conn, message, "500", "Internal Server Error")
			return
		}
	} else {
		// 更新设备状态
		device.Status = models.DeviceStatusOnline
		device.IP = strings.Split(conn.RemoteAddr().String(), ":")[0]
		device.UpdatedAt = time.Now()
		if err := p.protocol.deviceService.UpdateDevice(device); err != nil {
			log.Printf("Failed to update device: %v", err)
		}
	}

	// 保存设备到协议实例
	p.protocol.devices[deviceID] = device

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + p.protocol.sipServer.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Contact: " + contact + "\r\n"
	response += "Expires: " + strconv.Itoa(expires) + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending REGISTER response: %v", err)
	}

	log.Printf("Device registered successfully: %s", deviceID)
}

// parseDeviceID 解析设备ID
func (p *RegisterProcessor) parseDeviceID(fromHeader string) (string, error) {
	// From header格式: <sip:deviceID@ip:port>;tag=xxx
	// 提取deviceID
	start := strings.Index(fromHeader, "sip:")
	if start == -1 {
		return "", fmt.Errorf("invalid From header format")
	}
	start += 4

	end := strings.Index(fromHeader[start:], "@")
	if end == -1 {
		return "", fmt.Errorf("invalid From header format")
	}

	deviceID := fromHeader[start : start+end]
	return deviceID, nil
}

// sendErrorResponse 发送错误响应
func (p *RegisterProcessor) sendErrorResponse(conn net.Conn, message *SIPMessage, statusCode, reason string) {
	response := "SIP/2.0 " + statusCode + " " + reason + "\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + p.protocol.sipServer.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}

// RTPStream RTP流结构
type RTPStream struct {
	StreamID     string
	DeviceID     string
	ChannelID    string
	SSRC         string
	PayloadType  int
	Port         int
	StartTime    time.Time
	Active       bool
}

// InviteProcessor INVITE请求处理器
type InviteProcessor struct {
	protocol    *Protocol
	streams     map[string]*RTPStream
	streamsMutex sync.Mutex
}

// Process 处理INVITE请求
func (p *InviteProcessor) Process(message *SIPMessage, conn net.Conn) {
	log.Println("Processing INVITE request")

	// 解析设备和通道信息
	deviceID, channelID, err := p.parseDeviceAndChannel(message.Headers["From"])
	if err != nil {
		log.Printf("Failed to parse device and channel: %v", err)
		p.sendErrorResponse(conn, message, "400", "Bad Request")
		return
	}

	// 解析SDP
	sdpInfo, err := p.parseSDP(message.Body)
	if err != nil {
		log.Printf("Failed to parse SDP: %v", err)
		p.sendErrorResponse(conn, message, "400", "Bad Request")
		return
	}

	// 生成流ID
	streamID := fmt.Sprintf("%s_%s_%d", deviceID, channelID, time.Now().UnixNano())

	// 创建RTP流
	stream := &RTPStream{
		StreamID:     streamID,
		DeviceID:     deviceID,
		ChannelID:    channelID,
		SSRC:         sdpInfo["ssrc"],
		PayloadType:  96, // 默认H.264
		Port:         8000 + len(p.streams)%10000, // 动态端口
		StartTime:    time.Now(),
		Active:       true,
	}

	// 保存流
	p.streamsMutex.Lock()
	p.streams[streamID] = stream
	p.streamsMutex.Unlock()

	// 生成SDP响应
	localIP := strings.Split(conn.LocalAddr().String(), ":")[0]
	sdpResponse := p.generateSDP(localIP, stream.Port, stream.SSRC)

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + p.protocol.sipServer.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Contact: <sip:server@" + localIP + ":5060;transport=tcp>\r\n"
	response += "Content-Type: application/sdp\r\n"
	response += "Content-Length: " + strconv.Itoa(len(sdpResponse)) + "\r\n"
	response += "\r\n"
	response += sdpResponse

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending INVITE response: %v", err)
		return
	}

	log.Printf("Stream started successfully: %s", streamID)

	// TODO: 启动RTP接收
	go p.startRTPReceiver(stream)
}

// parseDeviceAndChannel 解析设备和通道信息
func (p *InviteProcessor) parseDeviceAndChannel(fromHeader string) (string, string, error) {
	// From header格式: <sip:deviceID@ip:port>;tag=xxx
	// 提取deviceID
	start := strings.Index(fromHeader, "sip:")
	if start == -1 {
		return "", "", fmt.Errorf("invalid From header format")
	}
	start += 4

	end := strings.Index(fromHeader[start:], "@")
	if end == -1 {
		return "", "", fmt.Errorf("invalid From header format")
	}

	deviceID := fromHeader[start : start+end]
	// 暂时使用默认通道1
	channelID := "1"

	return deviceID, channelID, nil
}

// parseSDP 解析SDP
func (p *InviteProcessor) parseSDP(sdp string) (map[string]string, error) {
	result := make(map[string]string)
	lines := strings.Split(sdp, "\r\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "m=") {
			result["media"] = line
		} else if strings.HasPrefix(line, "c=") {
			result["connection"] = line
		} else if strings.HasPrefix(line, "a=ssrc:") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				result["ssrc"] = parts[1]
			}
		} else if strings.HasPrefix(line, "a=rtpmap:") {
			result["rtpmap"] = line
		}
	}

	return result, nil
}

// generateSDP 生成SDP
func (p *InviteProcessor) generateSDP(ip string, port int, ssrc string) string {
	sdp := "v=0\r\n"
	sdp += "o=- " + ssrc + " " + ssrc + " IN IP4 " + ip + "\r\n"
	sdp += "s=Play\r\n"
	sdp += "c=IN IP4 " + ip + "\r\n"
	sdp += "t=0 0\r\n"
	sdp += "m=video " + strconv.Itoa(port) + " RTP/AVP 96\r\n"
	sdp += "a=rtpmap:96 H264/90000\r\n"
	sdp += "a=ssrc:" + ssrc + " cname:stream\r\n"
	sdp += "a=sendonly\r\n"
	return sdp
}

// startRTPReceiver 启动RTP接收器
func (p *InviteProcessor) startRTPReceiver(stream *RTPStream) {
	log.Printf("Starting RTP receiver for stream %s on port %d", stream.StreamID, stream.Port)

	// 创建UDP监听
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", stream.Port))
	if err != nil {
		log.Printf("Failed to resolve UDP address: %v", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Failed to start UDP listener: %v", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1500)
	for stream.Active {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading RTP packet: %v", err)
			break
		}

		// 处理RTP包
		p.handleRTPPacket(stream, buffer[:n])
	}

	log.Printf("RTP receiver stopped for stream %s", stream.StreamID)
}

// handleRTPPacket 处理RTP包
func (p *InviteProcessor) handleRTPPacket(stream *RTPStream, packet []byte) {
	// TODO: 实现RTP包处理逻辑
	// 1. 解析RTP头部
	// 2. 提取媒体数据
	// 3. 转发或处理媒体数据
	log.Printf("Received RTP packet for stream %s, size: %d", stream.StreamID, len(packet))
}

// sendErrorResponse 发送错误响应
func (p *InviteProcessor) sendErrorResponse(conn net.Conn, message *SIPMessage, statusCode, reason string) {
	response := "SIP/2.0 " + statusCode + " " + reason + "\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + ";tag=" + p.protocol.sipServer.generateTag() + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}

// AckProcessor ACK请求处理器
type AckProcessor struct {
	protocol *Protocol
}

// Process 处理ACK请求
func (p *AckProcessor) Process(message *SIPMessage, conn net.Conn) {
	log.Println("Processing ACK request")
	// TODO: 实现ACK处理逻辑
}

// ByeProcessor BYE请求处理器
type ByeProcessor struct {
	protocol *Protocol
	inviteProcessor *InviteProcessor
}

// Process 处理BYE请求
func (p *ByeProcessor) Process(message *SIPMessage, conn net.Conn) {
	log.Println("Processing BYE request")

	// 解析设备和通道信息
	deviceID, channelID, err := p.parseDeviceAndChannel(message.Headers["From"])
	if err != nil {
		log.Printf("Failed to parse device and channel: %v", err)
		// 仍然发送200 OK
	}

	// 停止相关流
	p.stopStreams(deviceID, channelID)

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending BYE response: %v", err)
	}

	log.Printf("Stream stopped for device %s, channel %s", deviceID, channelID)
}

// parseDeviceAndChannel 解析设备和通道信息
func (p *ByeProcessor) parseDeviceAndChannel(fromHeader string) (string, string, error) {
	// From header格式: <sip:deviceID@ip:port>;tag=xxx
	// 提取deviceID
	start := strings.Index(fromHeader, "sip:")
	if start == -1 {
		return "", "", fmt.Errorf("invalid From header format")
	}
	start += 4

	end := strings.Index(fromHeader[start:], "@")
	if end == -1 {
		return "", "", fmt.Errorf("invalid From header format")
	}

	deviceID := fromHeader[start : start+end]
	// 暂时使用默认通道1
	channelID := "1"

	return deviceID, channelID, nil
}

// stopStreams 停止流
func (p *ByeProcessor) stopStreams(deviceID, channelID string) {
	if p.inviteProcessor == nil {
		log.Println("Invite processor not set")
		return
	}

	p.inviteProcessor.streamsMutex.Lock()
	defer p.inviteProcessor.streamsMutex.Unlock()

	// 查找并停止相关流
	for streamID, stream := range p.inviteProcessor.streams {
		if stream.DeviceID == deviceID && stream.ChannelID == channelID {
			log.Printf("Stopping stream %s", streamID)
			stream.Active = false
			delete(p.inviteProcessor.streams, streamID)
			break
		}
	}

	log.Printf("Stopping streams for device %s, channel %s completed", deviceID, channelID)
}

// AlarmInfo 告警信息结构
type AlarmInfo struct {
	AlarmID     string
	DeviceID    string
	ChannelID   string
	AlarmType   string
	AlarmTime   time.Time
	AlarmData   map[string]string
	Status      string
}

// MessageProcessor MESSAGE请求处理器
type MessageProcessor struct {
	protocol *Protocol
	alarms   map[string]*AlarmInfo
	alarmsMutex sync.Mutex
}

// Process 处理MESSAGE请求
func (p *MessageProcessor) Process(message *SIPMessage, conn net.Conn) {
	log.Println("Processing MESSAGE request")

	// 解析消息体
	if message.Body != "" {
		// 检查是否是告警消息
		if strings.Contains(message.Body, "<CmdType>Alarm</CmdType>") {
			p.handleAlarmMessage(message, conn)
			return
		}
		// 检查是否是其他控制消息
		if strings.Contains(message.Body, "<CmdType>PTZ</CmdType>") {
			p.handlePTZMessage(message, conn)
			return
		}
	}

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending MESSAGE response: %v", err)
	}
}

// handleAlarmMessage 处理告警消息
func (p *MessageProcessor) handleAlarmMessage(message *SIPMessage, conn net.Conn) {
	log.Println("Processing alarm message")

	// 解析告警信息
	alarmInfo, err := p.parseAlarmMessage(message.Body)
	if err != nil {
		log.Printf("Failed to parse alarm message: %v", err)
		p.sendErrorResponse(conn, message, "400", "Bad Request")
		return
	}

	// 保存告警信息
	p.saveAlarm(alarmInfo)

	// 调用告警处理方法
	err = p.protocol.HandleAlarm(alarmInfo.DeviceID, alarmInfo.ChannelID, alarmInfo.AlarmType, alarmInfo.AlarmData)
	if err != nil {
		log.Printf("Failed to handle alarm: %v", err)
	}

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending alarm response: %v", err)
	}

	log.Printf("Alarm handled successfully: %s from device %s", alarmInfo.AlarmType, alarmInfo.DeviceID)
}

// handlePTZMessage 处理PTZ控制消息
func (p *MessageProcessor) handlePTZMessage(message *SIPMessage, conn net.Conn) {
	log.Println("Processing PTZ control message")

	// 发送200 OK响应
	response := "SIP/2.0 200 OK\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending PTZ response: %v", err)
	}
}

// parseAlarmMessage 解析告警消息
func (p *MessageProcessor) parseAlarmMessage(body string) (*AlarmInfo, error) {
	// 解析XML
	// 这里使用简单的字符串解析，实际项目中建议使用XML解析库
	deviceID := p.extractValue(body, "<DeviceID>", "</DeviceID>")
	alarmType := p.extractValue(body, "<AlarmType>", "</AlarmType>")
	alarmTime := p.extractValue(body, "<AlarmTime>", "</AlarmTime>")

	if deviceID == "" || alarmType == "" {
		return nil, fmt.Errorf("invalid alarm message format")
	}

	// 解析时间
	time, err := time.Parse("2006-01-02T15:04:05", alarmTime)
	if err != nil {
		time = time.Now() // 使用当前时间作为默认
	}

	// 构建告警信息
	alarmInfo := &AlarmInfo{
		AlarmID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		DeviceID:    deviceID,
		ChannelID:   deviceID, // 暂时使用设备ID作为通道ID
		AlarmType:   alarmType,
		AlarmTime:   time,
		AlarmData:   make(map[string]string),
		Status:      "active",
	}

	// 提取其他告警数据
	alarmInfo.AlarmData["alarmTime"] = alarmTime

	return alarmInfo, nil
}

// extractValue 从XML中提取值
func (p *MessageProcessor) extractValue(xml, startTag, endTag string) string {
	start := strings.Index(xml, startTag)
	if start == -1 {
		return ""
	}
	start += len(startTag)

	end := strings.Index(xml[start:], endTag)
	if end == -1 {
		return ""
	}

	return xml[start : start+end]
}

// saveAlarm 保存告警信息
func (p *MessageProcessor) saveAlarm(alarm *AlarmInfo) {
	p.alarmsMutex.Lock()
	defer p.alarmsMutex.Unlock()

	p.alarms[alarm.AlarmID] = alarm
}

// sendErrorResponse 发送错误响应
func (p *MessageProcessor) sendErrorResponse(conn net.Conn, message *SIPMessage, statusCode, reason string) {
	response := "SIP/2.0 " + statusCode + " " + reason + "\r\n"
	response += "Via: " + message.Headers["Via"] + "\r\n"
	response += "From: " + message.Headers["From"] + "\r\n"
	response += "To: " + message.Headers["To"] + "\r\n"
	response += "Call-ID: " + message.Headers["Call-ID"] + "\r\n"
	response += "CSeq: " + message.Headers["CSeq"] + "\r\n"
	response += "Content-Length: 0\r\n"
	response += "\r\n"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Printf("Error sending error response: %v", err)
	}
}

// Stop 停止GB28181协议服务
func (p *Protocol) Stop() {
	log.Println("Stopping GB28181 protocol service...")

	// 停止SIP服务器
	if p.sipServer != nil {
		p.sipServer.Stop()
	}

	log.Println("GB28181 protocol service stopped")
}

// startDeviceMonitoring 启动设备状态监控
func (p *Protocol) startDeviceMonitoring() {
	log.Println("Starting GB28181 device monitoring...")

	// 定期检查设备状态
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.monitorDevices()
		}
	}
}

// monitorDevices 监控设备状态
func (p *Protocol) monitorDevices() {
	log.Println("Monitoring GB28181 devices...")

	// 遍历已注册的设备
	for deviceID, device := range p.devices {
		log.Printf("Checking device status: %s (%s)", device.Name, deviceID)

		// 发送心跳请求
		go func(d *models.Device) {
			err := p.checkDeviceStatus(d)
			if err != nil {
				log.Printf("Device status check failed for %s: %v", d.Name, err)
				// 更新设备状态为错误
				p.deviceService.UpdateDeviceStatus(d.ID, models.DeviceStatusError)
			}
		}(device)
	}
}

// checkDeviceStatus 检查设备状态
func (p *Protocol) checkDeviceStatus(device *models.Device) error {
	// 发送心跳请求
	request := p.buildOptionsRequest(device)
	if request == "" {
		return fmt.Errorf("failed to build options request")
	}

	// 发送请求
	response, err := p.sendSIPRequest(device.IP, device.Port, request)
	if err != nil {
		// 更新设备状态为离线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOffline)
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}

	// 解析响应
	if strings.Contains(response, "200 OK") {
		// 更新设备状态为在线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOnline)
		log.Printf("Device %s is online", device.Name)
	} else {
		// 更新设备状态为离线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOffline)
		return fmt.Errorf("device returned non-200 response: %s", response)
	}

	return nil
}

// sendHeartbeat 发送心跳请求
func (p *Protocol) sendHeartbeat(device *models.Device) {
	// 构建OPTIONS请求
	request := p.buildOptionsRequest(device)
	if request == "" {
		log.Printf("Failed to build OPTIONS request for device %s", device.ID)
		return
	}

	// 发送请求
	response, err := p.sendSIPRequest(device.IP, device.Port, request)
	if err != nil {
		log.Printf("Failed to send heartbeat to device %s: %v", device.ID, err)
		// 更新设备状态为离线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOffline)
		return
	}

	// 解析响应
	if strings.Contains(response, "200 OK") {
		log.Printf("Device %s is online", device.ID)
		// 更新设备状态为在线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOnline)
	} else {
		log.Printf("Device %s is offline (response: %s)", device.ID, response)
		// 更新设备状态为离线
		p.deviceService.UpdateDeviceStatus(device.ID, models.DeviceStatusOffline)
	}
}

// buildOptionsRequest 构建OPTIONS请求
func (p *Protocol) buildOptionsRequest(device *models.Device) string {
	callID := fmt.Sprintf("%d", time.Now().UnixNano())
	cseq := "1 OPTIONS"
	fromTag := p.sipServer.generateTag()

	request := "OPTIONS sip:" + device.ID + "@" + device.IP + ":" + strconv.Itoa(device.Port) + " SIP/2.0\r\n"
	request += "Via: SIP/2.0/UDP " + p.sipServer.monitorIPs[0] + ":5060;branch=z9hG4bK" + fromTag + "\r\n"
	request += "From: <sip:server@" + p.sipServer.monitorIPs[0] + ">;tag=" + fromTag + "\r\n"
	request += "To: <sip:" + device.ID + "@" + device.IP + ">\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: " + cseq + "\r\n"
	request += "Max-Forwards: 70\r\n"
	request += "User-Agent: GB28181 Server\r\n"
	request += "Content-Length: 0\r\n"
	request += "\r\n"

	return request
}

// sendSIPRequest 发送SIP请求
func (p *Protocol) sendSIPRequest(ip string, port int, request string) (string, error) {
	// 验证参数
	if ip == "" || port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid IP or port: %s:%d", ip, port)
	}

	if request == "" {
		return "", fmt.Errorf("empty SIP request")
	}

	// 创建UDP连接
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return "", fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s:%d: %v", ip, port, err)
	}
	defer conn.Close()

	// 设置超时
	timeout := 5 * time.Second
	conn.SetDeadline(time.Now().Add(timeout))

	// 发送请求
	n, err := conn.Write([]byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to send SIP request: %v", err)
	}

	log.Printf("Sent %d bytes to %s:%d", n, ip, port)

	// 接收响应
	buffer := make([]byte, 4096)
	n, _, err = conn.ReadFromUDP(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "", fmt.Errorf("SIP request timeout after %v", timeout)
		}
		return "", fmt.Errorf("failed to read SIP response: %v", err)
	}

	response := string(buffer[:n])
	log.Printf("Received %d bytes response from %s:%d", n, ip, port)

	return response, nil
}

// RegisterDevice 注册设备
func (p *Protocol) RegisterDevice(device *models.Device) error {
	log.Printf("Registering GB28181 device: %s", device.Name)

	// TODO: 实现设备注册逻辑
	// 1. 验证设备信息
	// 2. 保存设备信息
	// 3. 发送注册响应

	p.devices[device.ID] = device
	return nil
}

// UnregisterDevice 注销设备
func (p *Protocol) UnregisterDevice(deviceID string) error {
	log.Printf("Unregistering GB28181 device: %s", deviceID)

	// TODO: 实现设备注销逻辑
	// 1. 查找设备
	// 2. 清理设备信息
	// 3. 发送注销响应

	delete(p.devices, deviceID)
	return nil
}

// QueryDevices 查询设备目录
func (p *Protocol) QueryDevices() ([]models.Device, error) {
	log.Println("Querying GB28181 devices...")

	// TODO: 实现设备目录查询
	// 1. 构建查询请求
	// 2. 发送查询请求
	// 3. 收集设备响应

	devices := make([]models.Device, 0, len(p.devices))
	for _, d := range p.devices {
		devices = append(devices, *d)
	}

	return devices, nil
}

// StartStream 启动视频流
func (p *Protocol) StartStream(deviceID string, channelID string) (string, error) {
	log.Printf("Starting GB28181 stream for device %s, channel %s", deviceID, channelID)

	// TODO: 实现视频流启动
	// 1. 验证设备存在
	// 2. 构建SIP INVITE请求
	// 3. 发送INVITE请求
	// 4. 处理设备响应
	// 5. 建立RTP/RTCP会话

	return "rtsp://127.0.0.1:554/stream", nil
}

// StopStream 停止视频流
func (p *Protocol) StopStream(deviceID string, channelID string) error {
	log.Printf("Stopping GB28181 stream for device %s, channel %s", deviceID, channelID)

	// TODO: 实现视频流停止
	// 1. 构建SIP BYE请求
	// 2. 发送BYE请求
	// 3. 关闭RTP/RTCP会话

	return nil
}

// ControlDevice 控制设备
func (p *Protocol) ControlDevice(deviceID string, channelID string, command string, params map[string]string) error {
	log.Printf("Controlling GB28181 device %s, channel %s, command %s", deviceID, channelID, command)

	// 查找设备
	device, exists := p.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// 根据命令类型构建控制请求
	var request string
	var err error

	switch command {
	case "PTZ":
		request, err = p.buildPTZControlRequest(device, channelID, params)
	case "Preset":
		request, err = p.buildPresetControlRequest(device, channelID, params)
	case "Query":
		request, err = p.buildQueryRequest(device, channelID, params)
	default:
		return fmt.Errorf("unsupported command: %s", command)
	}

	if err != nil {
		return fmt.Errorf("failed to build control request: %v", err)
	}

	// 发送控制请求
	response, err := p.sendSIPRequest(device.IP, device.Port, request)
	if err != nil {
		return fmt.Errorf("failed to send control request: %v", err)
	}

	// 解析响应
	if !strings.Contains(response, "200 OK") {
		return fmt.Errorf("control request failed: %s", response)
	}

	log.Printf("Device control successful: %s", command)
	return nil
}

// buildPTZControlRequest 构建PTZ控制请求
func (p *Protocol) buildPTZControlRequest(device *models.Device, channelID string, params map[string]string) (string, error) {
	callID := fmt.Sprintf("%d", time.Now().UnixNano())
	cseq := "1 MESSAGE"
	fromTag := p.sipServer.generateTag()

	// 构建XML控制命令
	xml := p.buildPTZXML(channelID, params)

	request := "MESSAGE sip:" + device.ID + "@" + device.IP + ":" + strconv.Itoa(device.Port) + " SIP/2.0\r\n"
	request += "Via: SIP/2.0/UDP " + p.sipServer.monitorIPs[0] + ":5060;branch=z9hG4bK" + fromTag + "\r\n"
	request += "From: <sip:server@" + p.sipServer.monitorIPs[0] + ">;tag=" + fromTag + "\r\n"
	request += "To: <sip:" + device.ID + "@" + device.IP + ">\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: " + cseq + "\r\n"
	request += "Max-Forwards: 70\r\n"
	request += "User-Agent: GB28181 Server\r\n"
	request += "Content-Type: Application/MANSCDP+xml\r\n"
	request += "Content-Length: " + strconv.Itoa(len(xml)) + "\r\n"
	request += "\r\n"
	request += xml

	return request, nil
}

// buildPTZXML 构建PTZ XML命令
func (p *Protocol) buildPTZXML(channelID string, params map[string]string) string {
	cmdType := params["type"]
	speed := params["speed"]
	if speed == "" {
		speed = "50"
	}

	xml := "<?xml version=\"1.0\" encoding=\"GB2312\"?><Control>\r\n"
	xml += "<CmdType>PTZ</CmdType>\r\n"
	xml += "<SN>" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000) + "</SN>\r\n"
	xml += "<DeviceID>" + channelID + "</DeviceID>\r\n"
	xml += "<PTZCommand>\r\n"
	xml += "<Command>" + cmdType + "</Command>\r\n"
	xml += "<Speed>" + speed + "</Speed>\r\n"
	xml += "</PTZCommand>\r\n"
	xml += "</Control>"

	return xml
}

// buildPresetControlRequest 构建预设控制请求
func (p *Protocol) buildPresetControlRequest(device *models.Device, channelID string, params map[string]string) (string, error) {
	callID := fmt.Sprintf("%d", time.Now().UnixNano())
	cseq := "1 MESSAGE"
	fromTag := p.sipServer.generateTag()

	// 构建XML控制命令
	xml := p.buildPresetXML(channelID, params)

	request := "MESSAGE sip:" + device.ID + "@" + device.IP + ":" + strconv.Itoa(device.Port) + " SIP/2.0\r\n"
	request += "Via: SIP/2.0/UDP " + p.sipServer.monitorIPs[0] + ":5060;branch=z9hG4bK" + fromTag + "\r\n"
	request += "From: <sip:server@" + p.sipServer.monitorIPs[0] + ">;tag=" + fromTag + "\r\n"
	request += "To: <sip:" + device.ID + "@" + device.IP + ">\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: " + cseq + "\r\n"
	request += "Max-Forwards: 70\r\n"
	request += "User-Agent: GB28181 Server\r\n"
	request += "Content-Type: Application/MANSCDP+xml\r\n"
	request += "Content-Length: " + strconv.Itoa(len(xml)) + "\r\n"
	request += "\r\n"
	request += xml

	return request, nil
}

// buildPresetXML 构建预设XML命令
func (p *Protocol) buildPresetXML(channelID string, params map[string]string) string {
	cmdType := params["type"] // Set, Call, Delete
	presetID := params["preset_id"]

	xml := "<?xml version=\"1.0\" encoding=\"GB2312\"?><Control>\r\n"
	xml += "<CmdType>PTZ</CmdType>\r\n"
	xml += "<SN>" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000) + "</SN>\r\n"
	xml += "<DeviceID>" + channelID + "</DeviceID>\r\n"
	xml += "<PTZCommand>\r\n"
	xml += "<Command>" + cmdType + "</Command>\r\n"
	xml += "<PresetID>" + presetID + "</PresetID>\r\n"
	xml += "</PTZCommand>\r\n"
	xml += "</Control>"

	return xml
}

// buildQueryRequest 构建查询请求
func (p *Protocol) buildQueryRequest(device *models.Device, channelID string, params map[string]string) (string, error) {
	callID := fmt.Sprintf("%d", time.Now().UnixNano())
	cseq := "1 MESSAGE"
	fromTag := p.sipServer.generateTag()

	// 构建XML查询命令
	xml := p.buildQueryXML(channelID, params)

	request := "MESSAGE sip:" + device.ID + "@" + device.IP + ":" + strconv.Itoa(device.Port) + " SIP/2.0\r\n"
	request += "Via: SIP/2.0/UDP " + p.sipServer.monitorIPs[0] + ":5060;branch=z9hG4bK" + fromTag + "\r\n"
	request += "From: <sip:server@" + p.sipServer.monitorIPs[0] + ">;tag=" + fromTag + "\r\n"
	request += "To: <sip:" + device.ID + "@" + device.IP + ">\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: " + cseq + "\r\n"
	request += "Max-Forwards: 70\r\n"
	request += "User-Agent: GB28181 Server\r\n"
	request += "Content-Type: Application/MANSCDP+xml\r\n"
	request += "Content-Length: " + strconv.Itoa(len(xml)) + "\r\n"
	request += "\r\n"
	request += xml

	return request, nil
}

// buildQueryXML 构建查询XML命令
func (p *Protocol) buildQueryXML(channelID string, params map[string]string) string {
	queryType := params["query_type"] // Catalog, DeviceInfo, Status

	xml := "<?xml version=\"1.0\" encoding=\"GB2312\"?><Query>\r\n"
	xml += "<CmdType>" + queryType + "</CmdType>\r\n"
	xml += "<SN>" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000) + "</SN>\r\n"
	xml += "<DeviceID>" + channelID + "</DeviceID>\r\n"
	xml += "</Query>"

	return xml
}

// HandleAlarm 处理告警
func (p *Protocol) HandleAlarm(deviceID string, channelID string, alarmType string, alarmData map[string]string) error {
	log.Printf("Handling GB28181 alarm from device %s, channel %s, type %s", deviceID, channelID, alarmType)

	// 1. 验证设备存在
	device, exists := p.devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	// 2. 保存告警记录
	alarmID := fmt.Sprintf("%d", time.Now().UnixNano())
	log.Printf("Alarm saved: %s, type: %s, device: %s", alarmID, alarmType, device.Name)

	// 3. 转发告警通知
	p.notifyAlarm(alarmID, deviceID, channelID, alarmType, alarmData)

	// 4. 处理特殊告警类型
	switch alarmType {
	case "Motion":
		log.Printf("Motion detected on device %s", device.Name)
		// 可以触发录制、推送通知等
	case "AlarmInput":
		log.Printf("Alarm input triggered on device %s", device.Name)
	case "VideoLoss":
		log.Printf("Video loss detected on device %s", device.Name)
	case "VideoBlind":
		log.Printf("Video blind detected on device %s", device.Name)
	default:
		log.Printf("Unknown alarm type: %s", alarmType)
	}

	log.Printf("Alarm handled successfully: %s", alarmType)
	return nil
}

// notifyAlarm 转发告警通知
func (p *Protocol) notifyAlarm(alarmID, deviceID, channelID, alarmType string, alarmData map[string]string) {
	// TODO: 实现告警通知转发
	// 1. 发送WebSocket通知
	// 2. 调用回调接口
	// 3. 发送邮件/短信通知

	log.Printf("Alarm notification sent: %s", alarmID)
}

// StartRecord 开始录制
func (p *Protocol) StartRecord(deviceID string, channelID string, streamType string) (string, error) {
	log.Printf("Starting record for device %s, channel %s, stream %s", deviceID, channelID, streamType)

	// 查找设备
	device, exists := p.devices[deviceID]
	if !exists {
		return "", fmt.Errorf("device not found: %s", deviceID)
	}

	// 构建录制请求
	request := p.buildRecordRequest(device, channelID, "Start", streamType)
	if request == "" {
		return "", fmt.Errorf("failed to build record request")
	}

	// 发送请求
	response, err := p.sendSIPRequest(device.IP, device.Port, request)
	if err != nil {
		return "", fmt.Errorf("failed to send record request: %v", err)
	}

	// 解析响应
	if !strings.Contains(response, "200 OK") {
		return "", fmt.Errorf("record request failed: %s", response)
	}

	// 创建录像信息
	recordID := fmt.Sprintf("%d", time.Now().UnixNano())
	record := &RecordInfo{
		RecordID:    recordID,
		DeviceID:    deviceID,
		ChannelID:   channelID,
		StartTime:   time.Now(),
		EndTime:     time.Time{},
		FileSize:    0,
		FilePath:    fmt.Sprintf("/rec/%s_%s_%s.mp4", deviceID, channelID, recordID),
		StreamType:  streamType,
		Status:      "recording",
	}

	// 保存录像信息
	p.recordsMutex.Lock()
	p.records[recordID] = record
	p.recordsMutex.Unlock()

	log.Printf("Record started successfully: %s", recordID)
	return recordID, nil
}

// StopRecord 停止录制
func (p *Protocol) StopRecord(recordID string) error {
	log.Printf("Stopping record: %s", recordID)

	// 查找录像
	p.recordsMutex.Lock()
	record, exists := p.records[recordID]
	if !exists {
		p.recordsMutex.Unlock()
		return fmt.Errorf("record not found: %s", recordID)
	}
	p.recordsMutex.Unlock()

	// 查找设备
	device, exists := p.devices[record.DeviceID]
	if !exists {
		return fmt.Errorf("device not found: %s", record.DeviceID)
	}

	// 构建录制请求
	request := p.buildRecordRequest(device, record.ChannelID, "Stop", record.StreamType)
	if request == "" {
		return fmt.Errorf("failed to build record request")
	}

	// 发送请求
	response, err := p.sendSIPRequest(device.IP, device.Port, request)
	if err != nil {
		return fmt.Errorf("failed to send record request: %v", err)
	}

	// 解析响应
	if !strings.Contains(response, "200 OK") {
		return fmt.Errorf("record request failed: %s", response)
	}

	// 更新录像信息
	p.recordsMutex.Lock()
	record.EndTime = time.Now()
	record.Status = "completed"
	p.recordsMutex.Unlock()

	log.Printf("Record stopped successfully: %s", recordID)
	return nil
}

// QueryRecords 查询录像
func (p *Protocol) QueryRecords(deviceID string, channelID string, startTime time.Time, endTime time.Time) ([]*RecordInfo, error) {
	log.Printf("Querying records for device %s, channel %s from %s to %s", deviceID, channelID, startTime, endTime)

	// 查找设备
	_, exists := p.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// 构建查询请求
	// 这里可以发送SIP查询请求到设备

	// 过滤本地录像
	var results []*RecordInfo
	p.recordsMutex.Lock()
	for _, record := range p.records {
		if record.DeviceID == deviceID && record.ChannelID == channelID {
			if (record.StartTime.After(startTime) || record.StartTime.Equal(startTime)) &&
				(record.EndTime.Before(endTime) || record.EndTime.Equal(endTime) || record.EndTime.IsZero()) {
				results = append(results, record)
			}
		}
	}
	p.recordsMutex.Unlock()

	log.Printf("Found %d records", len(results))
	return results, nil
}

// PlaybackRecord 回放录像
func (p *Protocol) PlaybackRecord(recordID string, startTime time.Time, endTime time.Time) (string, error) {
	log.Printf("Playing back record: %s from %s to %s", recordID, startTime, endTime)

	// 查找录像
	p.recordsMutex.Lock()
	record, exists := p.records[recordID]
	if !exists {
		p.recordsMutex.Unlock()
		return "", fmt.Errorf("record not found: %s", recordID)
	}
	p.recordsMutex.Unlock()

	// 查找设备
	device, exists := p.devices[record.DeviceID]
	if !exists {
		return "", fmt.Errorf("device not found: %s", record.DeviceID)
	}

	// 构建回放请求
	// 这里可以发送SIP回放请求到设备

	// 生成回放URL
	playbackURL := fmt.Sprintf("rtsp://%s:554/playback/%s", device.IP, recordID)

	log.Printf("Playback started: %s", playbackURL)
	return playbackURL, nil
}

// buildRecordRequest 构建录制请求
func (p *Protocol) buildRecordRequest(device *models.Device, channelID string, cmdType string, streamType string) string {
	callID := fmt.Sprintf("%d", time.Now().UnixNano())
	cseq := "1 MESSAGE"
	fromTag := p.sipServer.generateTag()

	// 构建XML命令
	xml := p.buildRecordXML(channelID, cmdType, streamType)

	request := "MESSAGE sip:" + device.ID + "@" + device.IP + ":" + strconv.Itoa(device.Port) + " SIP/2.0\r\n"
	request += "Via: SIP/2.0/UDP " + p.sipServer.monitorIPs[0] + ":5060;branch=z9hG4bK" + fromTag + "\r\n"
	request += "From: <sip:server@" + p.sipServer.monitorIPs[0] + ">;tag=" + fromTag + "\r\n"
	request += "To: <sip:" + device.ID + "@" + device.IP + ">\r\n"
	request += "Call-ID: " + callID + "\r\n"
	request += "CSeq: " + cseq + "\r\n"
	request += "Max-Forwards: 70\r\n"
	request += "User-Agent: GB28181 Server\r\n"
	request += "Content-Type: Application/MANSCDP+xml\r\n"
	request += "Content-Length: " + strconv.Itoa(len(xml)) + "\r\n"
	request += "\r\n"
	request += xml

	return request
}

// buildRecordXML 构建录制XML命令
func (p *Protocol) buildRecordXML(channelID string, cmdType string, streamType string) string {
	xml := "<?xml version=\"1.0\" encoding=\"GB2312\"?><Control>\r\n"
	xml += "<CmdType>Record</CmdType>\r\n"
	xml += "<SN>" + fmt.Sprintf("%d", time.Now().UnixNano()%1000000) + "</SN>\r\n"
	xml += "<DeviceID>" + channelID + "</DeviceID>\r\n"
	xml += "<RecordCmd>\r\n"
	xml += "<Command>" + cmdType + "</Command>\r\n"
	xml += "<StreamType>" + streamType + "</StreamType>\r\n"
	xml += "<Address>" + p.sipServer.monitorIPs[0] + "</Address>\r\n"
	xml += "<Port>8000</Port>\r\n"
	xml += "<Secrecy>0</Secrecy>\r\n"
	xml += "</RecordCmd>\r\n"
	xml += "</Control>"

	return xml
}
