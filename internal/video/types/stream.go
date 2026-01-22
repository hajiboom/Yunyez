package types

import (
	"time"
)

// StreamState represents the state of a stream
type StreamState string

const (
	StreamInactive StreamState = "inactive"
	StreamActive   StreamState = "active"
	StreamPaused   StreamState = "paused"
)

// MediaType represents the type of media
type MediaType string

const (
	VideoMediaType MediaType = "video"
	AudioMediaType MediaType = "audio"
)

// MediaFormat represents the format of media
type MediaFormat string

const (
	H264MediaFormat MediaFormat = "H264"
	H265MediaFormat MediaFormat = "H265"
	PCMAMediaFormat MediaFormat = "PCMA"
	PCMUFormat      MediaFormat = "PCMU"
)

// Stream represents a media stream
type Stream struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	State         StreamState            `json:"state"`
	MediaType     MediaType              `json:"media_type"`
	MediaFormat   MediaFormat            `json:"media_format"`
	Resolution    string                 `json:"resolution"` // e.g., "1920x1080"
	Bitrate       int                    `json:"bitrate"`    // in kbps
	Framerate     float32                `json:"framerate"` // frames per second
	CreatedAt     time.Time              `json:"created_at"`
	LastActivity  time.Time              `json:"last_activity"`
	Source        string                 `json:"source"` // source URL or device
	SDPInfo       string                 `json:"sdp_info"` // SDP description
	ExtraHeaders  map[string]interface{} `json:"extra_headers,omitempty"`
}

// Client represents a connected client
type Client struct {
	SessionID     string    `json:"session_id"`
	StreamID      string    `json:"stream_id"`
	ClientIP      string    `json:"client_ip"`
	UserAgent     string    `json:"user_agent"`
	Transport     string    `json:"transport"` // e.g., "RTP/UDP", "RTP/TCP"
	ConnectionTime time.Time `json:"connection_time"`
	LastActivity  time.Time `json:"last_activity"`
	BytesTransferred int64  `json:"bytes_transferred"`
	ControlPath   string    `json:"control_path"`
}

// StreamStats represents statistics for a stream
type StreamStats struct {
	StreamID         string `json:"stream_id"`
	TotalConnections int    `json:"total_connections"`
	ActiveClients    int    `json:"active_clients"`
	BytesSent        int64  `json:"bytes_sent"`
	BytesReceived    int64  `json:"bytes_received"`
	BitrateKbps      int    `json:"bitrate_kbps"`
	StartTime        time.Time `json:"start_time"`
	Uptime           time.Duration `json:"uptime"`
}

// ServerStats represents server-wide statistics
type ServerStats struct {
	Uptime          time.Duration `json:"uptime"`
	TotalSessions   int           `json:"total_sessions"`
	CurrentSessions int           `json:"current_sessions"`
	TotalStreams    int           `json:"total_streams"`
	ActiveStreams   int           `json:"active_streams"`
	BytesSent       int64         `json:"bytes_sent"`
	BytesReceived   int64         `json:"bytes_received"`
	StartTime       time.Time     `json:"start_time"`
}