// Package tts provides text-to-speech services.
package tts

import "context"


// Service convert text to speech realtime interface
type Service interface {
	// Synthesize convert text to speech realtime
	Synthesize(ctx context.Context, text string) ([]byte, error)
}

type ConfigReader interface {
	// GetString reads the TTS configuration as a string.
	GetString(key string) string
	// GetFloat64 reads the TTS configuration as a float64.
	GetFloat64(key string) float64
}
