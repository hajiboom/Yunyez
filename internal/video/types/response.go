package types

// APIResponse represents a generic API response
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthStatus represents the health status of the RTSP server
type HealthStatus struct {
	Status       string      `json:"status"`
	Timestamp    string      `json:"timestamp"`
	Uptime       string      `json:"uptime"`
	Version      string      `json:"version"`
	RTSPServer   RTSPInfo    `json:"rtsp_server"`
	RuntimeStats RuntimeInfo `json:"runtime_stats"`
}

// RTSPInfo contains RTSP server specific information
type RTSPInfo struct {
	Address           string `json:"address"`
	ActiveConnections int    `json:"active_connections"`
	TotalStreams      int    `json:"total_streams"`
	BytesTransferred  int64  `json:"bytes_transferred"`
}

// RuntimeInfo contains runtime statistics
type RuntimeInfo struct {
	Goroutines    int     `json:"goroutines"`
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
}

// StreamStatus represents the status of a specific stream
type StreamStatus struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	State          StreamState `json:"state"`
	Clients        int         `json:"clients"`
	BitrateKbps    int         `json:"bitrate_kbps"`
	Resolution     string      `json:"resolution"`
	Codec          string      `json:"codec"`
	CreatedAt      string      `json:"created_at"`
	LastActivity   string      `json:"last_activity"`
}

// ClientStatus represents the status of a connected client
type ClientStatus struct {
	SessionID        string `json:"session_id"`
	StreamID         string `json:"stream_id"`
	ClientIP         string `json:"client_ip"`
	Transport        string `json:"transport"`
	ConnectionTime   string `json:"connection_time"`
	BytesTransferred int64  `json:"bytes_transferred"`
}