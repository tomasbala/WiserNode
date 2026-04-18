package stream

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/wiserstream/gb28181-server/internal/config"
)

type ZLMClient struct {
	baseURL string
	secret  string
	client  *http.Client
}

func NewZLMClient(cfg *config.ZLMediaKitConfig) *ZLMClient {
	return &ZLMClient{
		baseURL: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.HTTPPort),
		secret:  cfg.Secret,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type ZLMResponse struct {
	Code int                    `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func (c *ZLMClient) doRequest(api string, params url.Values) (*ZLMResponse, error) {
	if params == nil {
		params = url.Values{}
	}
	if c.secret != "" {
		params.Set("secret", c.secret)
	}

	reqURL := fmt.Sprintf("%s/index/api/%s?%s", c.baseURL, api, params.Encode())

	resp, err := c.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result ZLMResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &result, nil
}

func (c *ZLMClient) GetMediaList(app string) (*ZLMResponse, error) {
	params := url.Values{}
	if app != "" {
		params.Set("app", app)
	}
	return c.doRequest("getMediaList", params)
}

func (c *ZLMClient) OpenRTPServer(streamID string, port int) (*ZLMResponse, error) {
	params := url.Values{}
	params.Set("stream_id", streamID)
	if port > 0 {
		params.Set("port", strconv.Itoa(port))
	}
	params.Set("enable_tcp", "1")
	params.Set("recreate", "1")
	return c.doRequest("openRtpServer", params)
}

func (c *ZLMClient) CloseRTPServer(streamID string) (*ZLMResponse, error) {
	params := url.Values{}
	params.Set("stream_id", streamID)
	return c.doRequest("closeRtpServer", params)
}

func (c *ZLMClient) GetRTPPort(streamID string) (int, error) {
	params := url.Values{}
	params.Set("stream_id", streamID)
	resp, err := c.doRequest("openRtpServer", params)
	if err != nil {
		return 0, err
	}

	if resp.Code != 0 {
		return 0, fmt.Errorf("open rtp server failed: %s", resp.Msg)
	}

	port, ok := resp.Data["port"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid port in response")
	}

	return int(port), nil
}

func (c *ZLMClient) StartSendRTP(streamID, ssrc string, dstHost string, dstPort int) (*ZLMResponse, error) {
	params := url.Values{}
	params.Set("vhost", "__defaultVhost__")
	params.Set("app", "rtp")
	params.Set("stream", streamID)
	params.Set("ssrc", ssrc)
	params.Set("dst_url", dstHost)
	params.Set("dst_port", strconv.Itoa(dstPort))
	return c.doRequest("startSendRtp", params)
}

func (c *ZLMClient) StopSendRTP(streamID string) (*ZLMResponse, error) {
	params := url.Values{}
	params.Set("vhost", "__defaultVhost__")
	params.Set("app", "rtp")
	params.Set("stream", streamID)
	return c.doRequest("stopSendRtp", params)
}

func (c *ZLMClient) GetStreamURL(streamID string, protocol string) string {
	switch protocol {
	case "rtsp":
		return fmt.Sprintf("rtsp://%s:554/rtp/%s", c.baseURL[7:], streamID)
	case "rtmp":
		return fmt.Sprintf("rtmp://%s/rtp/%s", c.baseURL[7:], streamID)
	case "flv":
		return fmt.Sprintf("%s/rtp/%s.live.flv", c.baseURL, streamID)
	case "hls":
		return fmt.Sprintf("%s/rtp/%s/hls.m3u8", c.baseURL, streamID)
	case "webrtc":
		return fmt.Sprintf("%s/index/api/webrtc?app=rtp&stream=%s", c.baseURL, streamID)
	default:
		return fmt.Sprintf("%s/rtp/%s.live.flv", c.baseURL, streamID)
	}
}

func (c *ZLMClient) GetServerConfig() (*ZLMResponse, error) {
	return c.doRequest("getServerConfig", nil)
}

func (c *ZLMClient) GetStatistic() (*ZLMResponse, error) {
	return c.doRequest("getStatistic", nil)
}
