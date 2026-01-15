// Package rtsp provides constants for RTSP protocol
package rtsp

// RTSP Methods
const (
	Options      = "OPTIONS"       // Get server capabilities
	Describe     = "DESCRIBE"      // Get media description
	Setup        = "SETUP"         // Establish media transport
	Play         = "PLAY"          // Start media playback
	Pause        = "PAUSE"         // Pause media playback
	Teardown     = "TEARDOWN"      // Release media transport
	GetParameter = "GET_PARAMETER" // Get parameter value
	SetParameter = "SET_PARAMETER" // Set parameter value
)

// RTSP Version
const (
	RTSPVersion = "RTSP/1.0" // RTSP version 1.0
)

// Common headers
const (
	CSeqHeader          = "CSeq"           // Sequence number for requests/responses
	ContentTypeHeader   = "Content-Type"   // Media type of request body
	ContentLengthHeader = "Content-Length" // Length of request body
	SessionHeader       = "Session"        // Session ID for media transport
	TransportHeader     = "Transport"      // Transport protocol and mode
	PublicHeader        = "Public"         // Supported RTSP methods
)

// Transport parameters
const (
	TransportRTPAVP    = "RTP/AVP"     // Transport protocol: RTP/AVP
	TransportUnicast   = "unicast"     // Transport mode: unicast
	TransportMulticast = "multicast"   // Transport mode: multicast
	ClientPortParam    = "client_port" // Client port range
	ServerPortParam    = "server_port" // Server port range
	SsrcParam          = "ssrc"        // Synchronization source identifier
)

// SDP content type
const (
	ApplicationSDP = "application/sdp" // Media type: SDP
)
