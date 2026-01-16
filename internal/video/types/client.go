// Package types defines data structures used in the video client
package types

import (
	"net"
	"time"
)

// RTPInfo holds information about RTP streams
type RTPInfo struct {
	URL     string `json:"url"`
	Seq     int    `json:"seq"`
	RTPTime int    `json:"rtptime"`
}

// TransportInfo holds transport parameters
type TransportInfo struct {
	Protocol     string `json:"protocol"`
	Delivery     string `json:"delivery"`
	Mode         string `json:"mode"`
	ClientPorts  string `json:"client_ports"` // e.g., "8000-8001"
	ServerPorts  string `json:"server_ports"` // e.g., "8002-8003"
	SSRC         string `json:"ssrc"`
	Source       string `json:"source"` // Multicast source
	Layers       int    `json:"layers"` // For layered encodings
	Profiles     string `json:"profiles"` // For RTP profiles
}

// RTSPSession holds session information
type RTSPSession struct {
	ID            string        `json:"id"`
	StreamID      string        `json:"stream_id"`
	ClientAddr    *net.UDPAddr  `json:"client_addr"`
	ServerAddr    *net.UDPAddr  `json:"server_addr"`
	TransportInfo *TransportInfo `json:"transport_info"`
	CreatedAt     time.Time     `json:"created_at"`
	LastActivity  time.Time     `json:"last_activity"`
	Timeout       time.Duration `json:"timeout"`
}

// RTSPRequest holds parsed RTSP request information
type RTSPRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Version     string            `json:"version"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	CSeq        int               `json:"cseq"`
	SessionID   string            `json:"session_id"`
	Transport   *TransportInfo    `json:"transport,omitempty"`
}

// RTSPResponse holds RTSP response information
type RTSPResponse struct {
	Version     string            `json:"version"`
	StatusCode  int               `json:"status_code"`
	StatusText  string            `json:"status_text"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	SessionID   string            `json:"session_id,omitempty"`
	RTPInfo     *RTPInfo          `json:"rtp_info,omitempty"`
}