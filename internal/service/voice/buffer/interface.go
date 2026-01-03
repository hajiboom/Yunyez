// Package buffer provides the text buffer for voice service.
// defined the audio buffer interface and methods
package buffer

import (
	mqttVoice "yunyez/internal/pkg/mqtt/protocol/voice"
)

type AudioBuffer interface {
	// Write writes audio data to the buffer
	Write(data []byte) (int, error)

	// Close closes the buffer
	Close() error

	SetClientID(clientID string)
	SetAudioFormat(format mqttVoice.AudioConfig)
}
