// Package tts provides text-to-speech services using Edge TTS and ChatTTS.
package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var defaultEdgeEndpoint = "http://localhost:8003/tts"

// EdgeTTSConfig Edge TTS 配置
type EdgeTTSConfig struct {
	Endpoint string // Edge TTS endpoint URL
	Voice    string // Voice to use for synthesis
	Rate     string // e.g., "+0%", "-10%"
	Pitch    string // e.g., "+0Hz", "-2Hz"
	Volume   string // e.g., "+0%", "+5%"
}

// EdgeTTS Edge TTS 客户端
type EdgeTTS struct {
	config EdgeTTSConfig
	client *http.Client
}

// NewEdgeTTS creates a new EdgeTTS instance with the given configuration.
func NewEdgeTTS(config EdgeTTSConfig) *EdgeTTS {
	if config.Endpoint == "" {
		config.Endpoint = defaultEdgeEndpoint
	}
	return &EdgeTTS{
		config: config,
		client: &http.Client{},
	}
}

// Synthesize synthesizes the given text into speech using Edge TTS.
func (e *EdgeTTS) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	// build post request
	reqBody := map[string]string{
		"text":   text,
		"voice":  e.config.Voice,
		"rate":   e.config.Rate,
		"pitch":  e.config.Pitch,
		"volume": e.config.Volume,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.config.Endpoint, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// send request
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return body, nil
}

// Close 关闭 HTTP 客户端（无操作）
func (e *EdgeTTS) Close() error {
	return nil
}

// ChatTTSConfig ChatTTS 配置
type ChatTTSConfig struct {
	Endpoint    string // ChatTTS endpoint URL
	Voice       string
	Rate        string // e.g., "+0%", "-10%"
	Pitch       string // e.g., "+0Hz", "-2Hz"
	Volume      string // e.g., "+0%", "+5%"
	Temperature string // e.g., "0.7"
}

// ChatTTS ChatTTS 客户端
type ChatTTS struct {
	config ChatTTSConfig
}

// NewChatTTS creates a new ChatTTS instance with the given configuration.
func NewChatTTS(config ChatTTSConfig) *ChatTTS {
	return &ChatTTS{
		config: config,
	}
}

// Synthesize 语音合成 - ChatTTS HTTP 方式
func (c *ChatTTS) Synthesize(ctx context.Context, text string) ([]byte, error) {
	// TODO: Implement ChatTTS synthesis
	return nil, nil
}

// Close 关闭 HTTP 客户端（无操作）
func (c *ChatTTS) Close() error {
	return nil
}
