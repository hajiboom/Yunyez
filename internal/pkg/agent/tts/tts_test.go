// Package tts Test
package tts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfig struct {
	mock.Mock
}

func (m *mockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *mockConfig) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

// Replace global config with mock (if your config supports injection)
// For simplicity, we test EdgeTTS directly.

// tts_test.go
func TestNewTTSClient_Edge(t *testing.T) {
	mockCfg := new(mockConfig)

	mockCfg.On("GetString", "tts.model").Return("edge")
	mockCfg.On("GetString", "tts.edge.endpoint").Return("http://localhost:8003/tts")
	mockCfg.On("GetString", "tts.edge.params.voice").Return("zh-CN-YunyeNeural")
	mockCfg.On("GetString", "tts.edge.params.rate").Return("+0%")
	mockCfg.On("GetString", "tts.edge.params.pitch").Return("+0Hz")
	mockCfg.On("GetString", "tts.edge.params.volume").Return("+0%")

	client, err := NewTTSClientFromConfig(mockCfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	mockCfg.AssertExpectations(t)
}

func TestEdgeTTS_Synthesize(t *testing.T) {
	// 模拟一个返回 MP3 数据的服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tts", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, world!", reqBody["text"])
		assert.Equal(t, "zh-CN-XiaoyiNeural", reqBody["voice"])

		// 返回模拟的 MP3 数据（可以是任意非空字节）
		w.Header().Set("Content-Type", "audio/mpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake-mp3-data")) // 模拟音频流
	}))
	defer mockServer.Close()

	// 创建 EdgeTTS 客户端
	client := &EdgeTTS{
		config: EdgeTTSConfig{
			Endpoint: mockServer.URL + "/tts",
			Voice:    "zh-CN-XiaoyiNeural",
			Rate:     "+0%",
			Pitch:    "+0Hz",
			Volume:   "+0%",
		},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// 调用 Synthesize
	ctx := context.Background()
	audio, err := client.Synthesize(ctx, "Hello, world!")
	assert.NoError(t, err)
	assert.NotNil(t, audio)
	assert.Equal(t, "fake-mp3-data", string(audio))
}

