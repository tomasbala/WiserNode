package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SIPMessage struct {
	Method      string
	StatusCode  int
	StatusText  string
	RequestURI  string
	Via         []string
	From        string
	To          string
	CallID      string
	CSeq        string
	Contact     string
	MaxForwards int
	UserAgent   string
	Expires     int
	Content     string
	ContentType string
	AuthHeader  string
	RemoteAddr  string
	RemotePort  int
	Headers     map[string]string
}

type SDPInfo struct {
	SessionName string
	IP          string
	Port        int
	SSRC        string
	Codecs      []Codec
	ReceiveOnly bool
}

type Codec struct {
	PayloadType int
	Name        string
	ClockRate   int
}

type SIPRequest struct {
	Method     string
	RequestURI string
	Headers    map[string]string
	Body       string
}

func ParseSIPMessage(data string) *SIPMessage {
	msg := &SIPMessage{
		Headers: make(map[string]string),
	}

	lines := strings.Split(data, "\r\n")
	if len(lines) == 0 {
		return msg
	}

	firstLine := strings.TrimSpace(lines[0])
	parts := strings.Fields(firstLine)

	if len(parts) >= 1 {
		if parts[0] == "SIP/2.0" {
			if len(parts) >= 3 {
				msg.StatusCode, _ = strconv.Atoi(parts[1])
				msg.StatusText = strings.Join(parts[2:], " ")
			}
		} else {
			msg.Method = parts[0]
			if len(parts) >= 2 {
				msg.RequestURI = parts[1]
			}
		}
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if i+1 < len(lines) {
				msg.Content = strings.Join(lines[i+1:], "\r\n")
			}
			break
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			headerName := strings.TrimSpace(line[:colonIdx])
			headerValue := strings.TrimSpace(line[colonIdx+1:])
			msg.Headers[headerName] = headerValue

			switch strings.ToLower(headerName) {
			case "via", "v":
				msg.Via = append(msg.Via, headerValue)
			case "from", "f":
				msg.From = headerValue
			case "to", "t":
				msg.To = headerValue
			case "call-id", "i":
				msg.CallID = headerValue
			case "cseq":
				msg.CSeq = headerValue
			case "contact", "m":
				msg.Contact = headerValue
			case "max-forwards":
				msg.MaxForwards, _ = strconv.Atoi(headerValue)
			case "user-agent":
				msg.UserAgent = headerValue
			case "expires":
				msg.Expires, _ = strconv.Atoi(headerValue)
			case "www-authenticate", "authorization":
				msg.AuthHeader = headerValue
			case "content-type", "c":
				msg.ContentType = headerValue
			}
		}
	}

	return msg
}

func ExtractDeviceID(header string) string {
	if idx := strings.Index(header, "sip:"); idx != -1 {
		rest := header[idx+4:]
		if end := strings.Index(rest, "@"); end != -1 {
			return rest[:end]
		}
		if end := strings.Index(rest, ">"); end != -1 {
			return rest[:end]
		}
		if end := strings.Index(rest, ";"); end != -1 {
			return rest[:end]
		}
	}
	return ""
}

func ExtractTag(header string) string {
	if idx := strings.Index(header, "tag="); idx != -1 {
		rest := header[idx+4:]
		if end := strings.Index(rest, ";"); end != -1 {
			return rest[:end]
		}
		if end := strings.Index(rest, ">"); end != -1 {
			return rest[:end]
		}
		return rest
	}
	return ""
}

func ExtractBranch(via string) string {
	if idx := strings.Index(via, "branch="); idx != -1 {
		rest := via[idx+7:]
		if end := strings.Index(rest, ";"); end != -1 {
			return rest[:end]
		}
		return rest
	}
	return ""
}

func GenerateTag() string {
	return strconv.FormatInt(time.Now().UnixNano()%1000000000, 16)
}

func GenerateBranch() string {
	return fmt.Sprintf("z9hG4bK%s", strconv.FormatInt(time.Now().UnixNano(), 16))
}

func GenerateCallID(host string) string {
	return fmt.Sprintf("%d@%s", time.Now().UnixNano(), host)
}

func GenerateSN() int {
	return int(time.Now().UnixNano() % 1000000000)
}

func ParseXMLValue(xml, tag string) string {
	startTag := "<" + tag + ">"
	endTag := "</" + tag + ">"

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

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

func ParseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05", s)
	return t
}
