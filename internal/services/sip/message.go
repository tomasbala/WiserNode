package sip

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
	"github.com/wiserstream/gb28181-server/internal/models"
)

type MessageBuilder struct {
	cfg *config.Config
}

func NewMessageBuilder(cfg *config.Config) *MessageBuilder {
	return &MessageBuilder{cfg: cfg}
}

func (b *MessageBuilder) BuildRegister(deviceID string, expires int, auth string) string {
	cseq := time.Now().Unix()
	callID := models.GenerateCallID(b.cfg.Server.Host)
	tag := models.GenerateTag()
	branch := models.GenerateBranch()

	var authLine string
	if auth != "" {
		authLine = fmt.Sprintf("Authorization: %s\r\n", auth)
	}

	return fmt.Sprintf(
		"REGISTER sip:%s SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP %s:%d;rport;branch=%s\r\n"+
			"From: <sip:%s@%s>;tag=%s\r\n"+
			"To: <sip:%s@%s>\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %d REGISTER\r\n"+
			"Contact: <sip:%s@%s:%d>\r\n"+
			"Max-Forwards: 70\r\n"+
			"User-Agent: GB28181-Server\r\n"+
			"Expires: %d\r\n"+
			"%s"+
			"Content-Length: 0\r\n\r\n",
		b.cfg.Server.Domain,
		b.cfg.Server.Host, b.cfg.Server.SIPPort, branch,
		deviceID, b.cfg.Server.Domain, tag,
		deviceID, b.cfg.Server.Domain,
		callID,
		cseq,
		deviceID, b.cfg.Server.Host, b.cfg.Server.SIPPort,
		expires,
		authLine,
	)
}

func (b *MessageBuilder) BuildMessage(deviceID string, body string, contentType string) string {
	cseq := time.Now().Unix()
	callID := models.GenerateCallID(b.cfg.Server.Host)
	tag := models.GenerateTag()
	branch := models.GenerateBranch()

	return fmt.Sprintf(
		"MESSAGE sip:%s@%s SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP %s:%d;rport;branch=%s\r\n"+
			"From: <sip:%s@%s>;tag=%s\r\n"+
			"To: <sip:%s@%s>\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %d MESSAGE\r\n"+
			"Max-Forwards: 70\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n\r\n%s",
		deviceID, b.cfg.Server.Domain,
		b.cfg.Server.Host, b.cfg.Server.SIPPort, branch,
		b.cfg.Server.ID, b.cfg.Server.Domain, tag,
		deviceID, b.cfg.Server.Domain,
		callID,
		cseq,
		contentType,
		len(body), body,
	)
}

func (b *MessageBuilder) BuildInvite(channelID string, ssrc string, rtpPort int, callID string, tag string) string {
	cseq := time.Now().Unix()
	branch := models.GenerateBranch()

	sdp := b.BuildSDP(ssrc, rtpPort)

	return fmt.Sprintf(
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
		channelID, b.cfg.Server.Domain,
		b.cfg.Server.Host, b.cfg.Server.SIPPort, branch,
		b.cfg.Server.ID, b.cfg.Server.Domain, tag,
		channelID, b.cfg.Server.Domain,
		callID,
		cseq,
		b.cfg.Server.ID, b.cfg.Server.Host, b.cfg.Server.SIPPort,
		len(sdp), sdp,
	)
}

func (b *MessageBuilder) BuildBye(channelID string, callID string, fromTag string, toTag string) string {
	cseq := time.Now().Unix()
	branch := models.GenerateBranch()

	return fmt.Sprintf(
		"BYE sip:%s@%s SIP/2.0\r\n"+
			"Via: SIP/2.0/UDP %s:%d;rport;branch=%s\r\n"+
			"From: <sip:%s@%s>;tag=%s\r\n"+
			"To: <sip:%s@%s>;tag=%s\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %d BYE\r\n"+
			"Max-Forwards: 70\r\n"+
			"Content-Length: 0\r\n\r\n",
		channelID, b.cfg.Server.Domain,
		b.cfg.Server.Host, b.cfg.Server.SIPPort, branch,
		b.cfg.Server.ID, b.cfg.Server.Domain, fromTag,
		channelID, b.cfg.Server.Domain, toTag,
		callID,
		cseq,
	)
}

func (b *MessageBuilder) BuildOK(originalMsg string, toTag string) string {
	msg := models.ParseSIPMessage(originalMsg)

	via := ""
	if len(msg.Via) > 0 {
		via = fmt.Sprintf("Via: %s\r\n", msg.Via[0])
	}

	cseq := msg.CSeq
	if cseq != "" && !strings.Contains(cseq, " ") {
		cseq = cseq + " " + msg.Method
	}

	to := msg.To
	if toTag != "" && !strings.Contains(to, "tag=") {
		to = fmt.Sprintf("%s;tag=%s", to, toTag)
	}

	return fmt.Sprintf(
		"SIP/2.0 200 OK\r\n"+
			"%s"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %s\r\n"+
			"Content-Length: 0\r\n\r\n",
		via,
		msg.From,
		to,
		msg.CallID,
		cseq,
	)
}

func (b *MessageBuilder) BuildUnauthorized(originalMsg string, nonce string) string {
	msg := models.ParseSIPMessage(originalMsg)

	via := ""
	if len(msg.Via) > 0 {
		via = fmt.Sprintf("Via: %s\r\n", msg.Via[0])
	}

	cseq := msg.CSeq
	if cseq != "" && !strings.Contains(cseq, " ") {
		cseq = cseq + " " + msg.Method
	}

	return fmt.Sprintf(
		"SIP/2.0 401 Unauthorized\r\n"+
			"%s"+
			"From: %s\r\n"+
			"To: %s;tag=%s\r\n"+
			"Call-ID: %s\r\n"+
			"CSeq: %s\r\n"+
			"WWW-Authenticate: %s\r\n"+
			"Content-Length: 0\r\n\r\n",
		via,
		msg.From,
		msg.To, models.GenerateTag(),
		msg.CallID,
		cseq,
		BuildWWWAuthenticate(b.cfg.Server.Domain, nonce),
	)
}

func (b *MessageBuilder) BuildSDP(ssrc string, rtpPort int) string {
	sessionID := time.Now().Unix()

	return fmt.Sprintf(
		"v=0\r\n"+
			"o=- %d 0 IN IP4 %s\r\n"+
			"s=Play\r\n"+
			"c=IN IP4 %s\r\n"+
			"t=0 0\r\n"+
			"m=video %d RTP/AVP 96\r\n"+
			"a=rtpmap:96 PS/90000\r\n"+
			"a=recvonly\r\n"+
			"y=%s\r\n"+
			"f=v/2/4///a///\r\n",
		sessionID, b.cfg.Server.Host,
		b.cfg.Server.Host,
		rtpPort,
		ssrc,
	)
}

func (b *MessageBuilder) BuildSDPPlayback(ssrc string, rtpPort int, startTime, endTime string) string {
	sessionID := time.Now().Unix()

	return fmt.Sprintf(
		"v=0\r\n"+
			"o=- %d 0 IN IP4 %s\r\n"+
			"s=Playback\r\n"+
			"c=IN IP4 %s\r\n"+
			"t=0 0\r\n"+
			"m=video %d RTP/AVP 96\r\n"+
			"a=rtpmap:96 PS/90000\r\n"+
			"a=recvonly\r\n"+
			"a=startTime:%s\r\n"+
			"a=endTime:%s\r\n"+
			"y=%s\r\n"+
			"f=v/2/4///a///\r\n",
		sessionID, b.cfg.Server.Host,
		b.cfg.Server.Host,
		rtpPort,
		startTime, endTime,
		ssrc,
	)
}

func (b *MessageBuilder) BuildCatalogQuery(deviceID string) string {
	sn := models.GenerateSN()
	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>Catalog</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
</Query>`,
		sn, deviceID,
	)

	return b.BuildMessage(deviceID, body, "Application/MANSCDP+xml")
}

func (b *MessageBuilder) BuildRecordQuery(deviceID, channelID, startTime, endTime string, sn int) string {
	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Query>
<CmdType>RecordInfo</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<StartTime>%s</StartTime>
<EndTime>%s</EndTime>
<Secrecy>0</Secrecy>
<Type>all</Type>
</Query>`,
		sn, channelID, startTime, endTime,
	)

	return b.BuildMessage(deviceID, body, "Application/MANSCDP+xml")
}

func (b *MessageBuilder) BuildPTZControl(deviceID, channelID string, ptzCmd string) string {
	sn := models.GenerateSN()
	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<PTZCmd>%s</PTZCmd>
</Control>`,
		sn, channelID, ptzCmd,
	)

	return b.BuildMessage(deviceID, body, "Application/MANSCDP+xml")
}

func (b *MessageBuilder) BuildDeviceControl(deviceID, channelID, cmd string, param string) string {
	sn := models.GenerateSN()
	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Control>
<CmdType>DeviceControl</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
%s
</Control>`,
		sn, channelID, param,
	)

	return b.BuildMessage(deviceID, body, "Application/MANSCDP+xml")
}

func (b *MessageBuilder) BuildNotifyKeepalive(deviceID string) string {
	sn := models.GenerateSN()
	body := fmt.Sprintf(
		`<?xml version="1.0" encoding="GB2312"?>
<Notify>
<CmdType>Keepalive</CmdType>
<SN>%d</SN>
<DeviceID>%s</DeviceID>
<Status>OK</Status>
</Notify>`,
		sn, deviceID,
	)

	return b.BuildMessage(deviceID, body, "Application/MANSCDP+xml")
}

func (b *MessageBuilder) GeneratePTZCommand(action string, speed int) string {
	if speed <= 0 {
		speed = 5
	}
	if speed > 15 {
		speed = 15
	}

	var cmd string
	switch action {
	case models.PTZStop:
		cmd = "A50F0100"
	case models.PTZUp:
		cmd = fmt.Sprintf("A50F01%02X", 0x10|speed)
	case models.PTZDown:
		cmd = fmt.Sprintf("A50F01%02X", 0x20|speed)
	case models.PTZLeft:
		cmd = fmt.Sprintf("A50F01%02X", 0x40|speed)
	case models.PTZRight:
		cmd = fmt.Sprintf("A50F01%02X", 0x80|speed)
	case models.PTZZoomIn:
		cmd = fmt.Sprintf("A50F01%02X", 0x01|(speed<<4))
	case models.PTZZoomOut:
		cmd = fmt.Sprintf("A50F01%02X", 0x02|(speed<<4))
	case models.PTZFocusNear:
		cmd = fmt.Sprintf("A50F01%02X", 0x40|speed)
	case models.PTZFocusFar:
		cmd = fmt.Sprintf("A50F01%02X", 0x80|speed)
	default:
		cmd = "A50F0100"
	}

	checksum := b.calculatePTZChecksum(cmd)
	return cmd + "0000" + checksum
}

func (b *MessageBuilder) calculatePTZChecksum(cmd string) string {
	sum := 0
	for i := 0; i < len(cmd); i += 2 {
		b, _ := strconv.ParseInt(cmd[i:i+2], 16, 32)
		sum += int(b)
	}
	return fmt.Sprintf("%02X", (sum&0xFF)^0xFF+1)
}
