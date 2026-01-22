package video

import (
	"testing"
	"time"

	"yunyez/internal/video/types"
)

func TestStreamManager(t *testing.T) {
	sm := NewStreamManager()

	// Create a test stream
	stream := &types.Stream{
		ID:          "test-stream",
		Name:        "Test Stream",
		State:       types.StreamActive,
		MediaType:   types.VideoMediaType,
		MediaFormat: types.H264MediaFormat,
		Resolution:  "1920x1080",
		Bitrate:     2048,
		Framerate:   30.0,
		CreatedAt:   time.Now(),
		LastActivity: time.Now(),
		Source:      "test://source",
	}

	// Test adding a stream
	if err := sm.AddStream(stream); err != nil {
		t.Errorf("Failed to add stream: %v", err)
	}

	// Test getting the stream
	retrievedStream := sm.GetStream("test-stream")
	if retrievedStream == nil {
		t.Error("Failed to retrieve stream")
	} else if retrievedStream.ID != "test-stream" {
		t.Errorf("Expected stream ID 'test-stream', got '%s'", retrievedStream.ID)
	}

	// Test getting all streams
	allStreams := sm.GetStreams()
	if len(allStreams) != 1 {
		t.Errorf("Expected 1 stream, got %d", len(allStreams))
	}

	// Test updating a stream
	stream.Bitrate = 4096
	if err := sm.UpdateStream(stream); err != nil {
		t.Errorf("Failed to update stream: %v", err)
	}

	updatedStream := sm.GetStream("test-stream")
	if updatedStream.Bitrate != 4096 {
		t.Errorf("Expected bitrate 4096, got %d", updatedStream.Bitrate)
	}

	// Test removing a stream
	if err := sm.RemoveStream("test-stream"); err != nil {
		t.Errorf("Failed to remove stream: %v", err)
	}

	// Verify stream is removed
	removedStream := sm.GetStream("test-stream")
	if removedStream != nil {
		t.Error("Stream should be removed but still exists")
	}
}

func TestSDPGenerator(t *testing.T) {
	generator := NewSDPGenerator()

	stream := &types.Stream{
		ID:          "test-stream",
		Name:        "Test Stream",
		State:       types.StreamActive,
		MediaType:   types.VideoMediaType,
		MediaFormat: types.H264MediaFormat,
		Resolution:  "1920x1080",
		Bitrate:     2048,
		Framerate:   30.0,
		CreatedAt:   time.Now(),
		LastActivity: time.Now(),
		Source:      "test://source",
	}

	// Test generating SDP string
	sdpStr, err := generator.GenerateSDPString(stream)
	if err != nil {
		t.Errorf("Failed to generate SDP string: %v", err)
	}

	if sdpStr == "" {
		t.Error("Generated SDP string is empty")
	}

	// Verify SDP contains expected elements
	if !contains(sdpStr, "v=0") {
		t.Error("SDP string does not contain version")
	}

	if !contains(sdpStr, "m=video") {
		t.Error("SDP string does not contain video media description")
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return containsSubstring(str, substr)
}

// Helper function to check if a string contains a substring anywhere
func containsSubstring(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}