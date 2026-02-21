package test

import (
	"testing"

	"video-platform/internal/models"
	"video-platform/internal/services/stream"
)

func TestStreamService(t *testing.T) {
	// 创建流媒体服务实例
	service := stream.NewService()

	// 测试用例1: 启动流
	t.Run("StartStream", func(t *testing.T) {
		req := &models.StartStreamRequest{
			DeviceID:  "device-1",
			ChannelID: "1",
			Protocol:  "RTSP",
		}

		stream, err := service.StartStream(req)
		if err != nil {
			t.Errorf("StartStream failed: %v", err)
			return
		}

		if stream.ID == "" {
			t.Error("stream ID should not be empty")
		}

		if stream.DeviceID != req.DeviceID {
			t.Errorf("expected device ID %s, got %s", req.DeviceID, stream.DeviceID)
		}

		if stream.URL == "" {
			t.Error("stream URL should not be empty")
		}

		t.Logf("StartStream test passed, stream ID: %s, URL: %s", stream.ID, stream.URL)
	})

	// 测试用例2: 获取流列表
	t.Run("GetStreams", func(t *testing.T) {
		streams, err := service.GetStreams()
		if err != nil {
			t.Errorf("GetStreams failed: %v", err)
			return
		}

		if len(streams) == 0 {
			t.Error("stream list should not be empty")
		}

		t.Logf("GetStreams test passed, stream count: %d", len(streams))
	})

	// 测试用例3: 停止流
	t.Run("StopStream", func(t *testing.T) {
		// 先启动一个流
		req := &models.StartStreamRequest{
			DeviceID:  "device-2",
			ChannelID: "1",
		}

		newStream, err := service.StartStream(req)
		if err != nil {
			t.Errorf("StartStream failed: %v", err)
			return
		}

		// 停止流
		err = service.StopStream(newStream.ID)
		if err != nil {
			t.Errorf("StopStream failed: %v", err)
			return
		}

		// 获取流状态
		status, err := service.GetStreamStatus(newStream.ID)
		if err != nil {
			t.Errorf("GetStreamStatus failed: %v", err)
			return
		}

		if status != models.StreamStatusStopped {
			t.Errorf("expected stream status %s, got %s", models.StreamStatusStopped, status)
		}

		t.Logf("StopStream test passed, stream stopped: %s", newStream.ID)
	})

	// 测试用例4: 获取流信息
	t.Run("GetStream", func(t *testing.T) {
		// 先启动一个流
		req := &models.StartStreamRequest{
			DeviceID:  "device-3",
			ChannelID: "1",
		}

		newStream, err := service.StartStream(req)
		if err != nil {
			t.Errorf("StartStream failed: %v", err)
			return
		}

		// 获取流信息
		stream, err := service.GetStream(newStream.ID)
		if err != nil {
			t.Errorf("GetStream failed: %v", err)
			return
		}

		if stream.ID != newStream.ID {
			t.Errorf("expected stream ID %s, got %s", newStream.ID, stream.ID)
		}

		t.Logf("GetStream test passed, stream: %s", stream.Name)
	})

	// 测试用例5: 获取流URL
	t.Run("GetStreamURL", func(t *testing.T) {
		// 先启动一个流
		req := &models.StartStreamRequest{
			DeviceID:  "device-4",
			ChannelID: "1",
		}

		newStream, err := service.StartStream(req)
		if err != nil {
			t.Errorf("StartStream failed: %v", err)
			return
		}

		// 获取流URL
		url, err := service.GetStreamURL(newStream.ID)
		if err != nil {
			t.Errorf("GetStreamURL failed: %v", err)
			return
		}

		if url == "" {
			t.Error("stream URL should not be empty")
		}

		t.Logf("GetStreamURL test passed, URL: %s", url)
	})

	// 测试用例6: 重启流
	t.Run("RestartStream", func(t *testing.T) {
		// 先启动一个流
		req := &models.StartStreamRequest{
			DeviceID:  "device-5",
			ChannelID: "1",
		}

		newStream, err := service.StartStream(req)
		if err != nil {
			t.Errorf("StartStream failed: %v", err)
			return
		}

		// 重启流
		err = service.RestartStream(newStream.ID)
		if err != nil {
			t.Errorf("RestartStream failed: %v", err)
			return
		}

		// 获取流状态
		status, err := service.GetStreamStatus(newStream.ID)
		if err != nil {
			t.Errorf("GetStreamStatus failed: %v", err)
			return
		}

		if status != models.StreamStatusPlaying {
			t.Errorf("expected stream status %s, got %s", models.StreamStatusPlaying, status)
		}

		t.Logf("RestartStream test passed, stream restarted: %s", newStream.ID)
	})
}
