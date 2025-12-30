// Package tts provides text-to-speech services using Edge TTS.
package tts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var defaultEdgeEndpoint = "https://localhost:8003/tts"

type EdgeTTSConfig struct {
	Endpoint string // Edge TTS endpoint URL
	Voice string // Voice to use for synthesis
	Rate    string // e.g., "+0%", "-10%"
	Pitch   string // e.g., "+0Hz", "-2Hz"
	Volume  string // e.g., "+0%", "+5%"
}


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
// the edge-tts server is a fastapi server so that we need to do a post request to the server
// Parameters:
//   - ctx: Context for request cancellation and timeout.
//   - text: The text to be synthesized into speech.
// Returns:
//   - []byte: The synthesized audio data.
//   - error: Any error encountered during the synthesis process.
func (e *EdgeTTS) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	// build post request
	reqBody := map[string]string{
		"text": text,
		"voice": e.config.Voice,
		"rate": e.config.Rate,
		"pitch": e.config.Pitch,
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