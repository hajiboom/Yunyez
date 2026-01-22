package video

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"yunyez/internal/pkg/media/sdp"
	"yunyez/internal/pkg/rtsp"
	"yunyez/internal/video/interfaces"
	"yunyez/internal/video/types"
)

// RTSPHandler handles RTSP requests
type RTSPHandler struct {
	streamManager     interfaces.StreamManager
	connectionManager interfaces.ConnectionManager
}

// NewRTSPHandler creates a new RTSP handler
func NewRTSPHandler(streamMgr interfaces.StreamManager, connMgr interfaces.ConnectionManager) *RTSPHandler {
	return &RTSPHandler{
		streamManager:     streamMgr,
		connectionManager: connMgr,
	}
}

// HandleRequest handles an incoming RTSP request
func (h *RTSPHandler) HandleRequest(conn net.Conn, req *types.RTSPRequest) (*types.RTSPResponse, error) {
	switch req.Method {
	case rtsp.Options:
		return h.handleOptions(req)
	case rtsp.Describe:
		return h.handleDescribe(req)
	case rtsp.Setup:
		return h.handleSetup(conn, req)
	case rtsp.Play:
		return h.handlePlay(req)
	case rtsp.Pause:
		return h.handlePause(req)
	case rtsp.Teardown:
		return h.handleTeardown(req)
	default:
		return h.createErrorResponse(501, "Method Not Implemented", req.CSeq)
	}
}

// handleOptions handles OPTIONS requests
func (h *RTSPHandler) handleOptions(req *types.RTSPRequest) (*types.RTSPResponse, error) {
	headers := map[string]string{
		rtsp.CSeqHeader:   strconv.Itoa(req.CSeq),
		rtsp.PublicHeader: rtsp.FormatSupportedMethods(),
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
	}, nil
}

// handleDescribe handles DESCRIBE requests
func (h *RTSPHandler) handleDescribe(req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Parse the URL to extract stream ID
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return h.createErrorResponse(400, "Bad Request", req.CSeq)
	}

	streamID := strings.TrimPrefix(parsedURL.Path, "/")
	if streamID == "" {
		// If no stream ID in path, try to get from host
		streamID = parsedURL.Host
		if streamID == "" {
			return h.createErrorResponse(400, "Bad Request", req.CSeq)
		}
	}

	// Get the stream
	stream := h.streamManager.GetStream(streamID)
	if stream == nil {
		return h.createErrorResponse(404, "Stream Not Found", req.CSeq)
	}

	// Generate SDP description
	sdpDesc := h.generateSDPForStream(stream)
	sdpStr := sdpDesc.String()

	headers := map[string]string{
		rtsp.CSeqHeader:          strconv.Itoa(req.CSeq),
		rtsp.ContentTypeHeader:   rtsp.ApplicationSDP,
		rtsp.ContentLengthHeader: strconv.Itoa(len(sdpStr)),
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
		Body:       []byte(sdpStr),
	}, nil
}

// handleSetup handles SETUP requests
func (h *RTSPHandler) handleSetup(conn net.Conn, req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Parse the URL to extract stream and track IDs
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return h.createErrorResponse(400, "Bad Request", req.CSeq)
	}

	streamPath := strings.TrimPrefix(parsedURL.Path, "/")
	parts := strings.Split(streamPath, "/")
	if len(parts) == 0 {
		return h.createErrorResponse(400, "Bad Request", req.CSeq)
	}

	streamID := parts[0]
	var trackID string
	if len(parts) > 1 {
		trackID = parts[1]
	} else {
		trackID = "track1" // Default track ID
	}
	_ = trackID // Use the variable to avoid "declared but not used" error

	// Check if stream exists
	stream := h.streamManager.GetStream(streamID)
	if stream == nil {
		return h.createErrorResponse(404, "Stream Not Found", req.CSeq)
	}

	// Parse transport header
	transportStr, exists := req.Headers[rtsp.TransportHeader]
	if !exists {
		return h.createErrorResponse(461, "Unsupported Transport", req.CSeq)
	}

	transportInfo, err := rtsp.ParseTransportHeader(transportStr)
	if err != nil {
		return h.createErrorResponse(461, "Unsupported Transport", req.CSeq)
	}

	// Generate session ID if not provided
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID, err = rtsp.GenerateSessionID()
		if err != nil {
			return h.createErrorResponse(500, "Internal Server Error", req.CSeq)
		}
	}

	// Create transport info
	transInfo := &types.TransportInfo{
		Protocol:    transportInfo["transport"],
		Delivery:    transportInfo["delivery"],
		Mode:        transportInfo["mode"],
		ClientPorts: transportInfo[rtsp.ClientPortParam],
		SSRC:        transportInfo[rtsp.SsrcParam],
	}

	// Assign server ports (in a real implementation, these would be dynamically assigned)
	if transInfo.Protocol == rtsp.TransportRTPAVP {
		if transInfo.Delivery == rtsp.TransportUnicast {
			// For unicast, assign server ports (typically consecutive even/odd for RTP/RTCP)
			transInfo.ServerPorts = "8002-8003" // Example ports
		}
	}

	// Create RTSP session
	clientAddr := conn.RemoteAddr().(*net.TCPAddr)
	_ = &types.RTSPSession{
		ID:            sessionID,
		StreamID:      streamID,
		ClientAddr:    &net.UDPAddr{IP: clientAddr.IP, Port: clientAddr.Port},
		TransportInfo: transInfo,
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
		Timeout:       60 * time.Second, // Default timeout
	}
	// Note: session is not currently used but kept for future implementation

	// Create client record
	client := &types.Client{
		SessionID:      sessionID,
		StreamID:       streamID,
		ClientIP:       clientAddr.IP.String(),
		Transport:      fmt.Sprintf("%s/%s", transInfo.Protocol, transInfo.Delivery),
		ConnectionTime: time.Now(),
		LastActivity:   time.Now(),
		ControlPath:    req.URL,
	}

	// Add client to connection manager
	err = h.connectionManager.AddClient(client)
	if err != nil {
		return h.createErrorResponse(500, "Internal Server Error", req.CSeq)
	}

	// Prepare response headers
	headers := map[string]string{
		rtsp.CSeqHeader:    strconv.Itoa(req.CSeq),
		rtsp.SessionHeader: sessionID + ";timeout=60",
		rtsp.TransportHeader: rtsp.FormatTransportHeader(map[string]string{
			"transport":          transInfo.Protocol,
			"delivery":           transInfo.Delivery,
			"source":             clientAddr.IP.String(),
			rtsp.ServerPortParam: transInfo.ServerPorts,
			rtsp.SsrcParam:       transInfo.SSRC,
		}),
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
		SessionID:  sessionID,
	}, nil
}

// handlePlay handles PLAY requests
func (h *RTSPHandler) handlePlay(req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Verify session exists
	if req.SessionID == "" {
		return h.createErrorResponse(454, "Session Not Found", req.CSeq)
	}

	// Get client by session ID
	client := h.connectionManager.GetClient(req.SessionID)
	if client == nil {
		return h.createErrorResponse(454, "Session Not Found", req.CSeq)
	}

	// Update client activity
	client.LastActivity = time.Now()
	h.connectionManager.UpdateClient(client)

	// Update stream state to active if it's not already
	stream := h.streamManager.GetStream(client.StreamID)
	if stream != nil && stream.State != types.StreamActive {
		stream.State = types.StreamActive
		h.streamManager.UpdateStream(stream)
	}

	// Prepare response headers
	headers := map[string]string{
		rtsp.CSeqHeader:    strconv.Itoa(req.CSeq),
		rtsp.SessionHeader: req.SessionID,
	}

	// Add RTP-Info header if available
	rtpInfo := &types.RTPInfo{
		URL:     req.URL,
		Seq:     0, // Starting sequence number
		RTPTime: 0, // Starting RTP timestamp
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
		SessionID:  req.SessionID,
		RTPInfo:    rtpInfo,
	}, nil
}

// handlePause handles PAUSE requests
func (h *RTSPHandler) handlePause(req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Verify session exists
	if req.SessionID == "" {
		return h.createErrorResponse(454, "Session Not Found", req.CSeq)
	}

	// Get client by session ID
	client := h.connectionManager.GetClient(req.SessionID)
	if client == nil {
		return h.createErrorResponse(454, "Session Not Found", req.CSeq)
	}

	// Update client activity
	client.LastActivity = time.Now()
	h.connectionManager.UpdateClient(client)

	// Update stream state to paused
	stream := h.streamManager.GetStream(client.StreamID)
	if stream != nil {
		stream.State = types.StreamPaused
		h.streamManager.UpdateStream(stream)
	}

	// Prepare response headers
	headers := map[string]string{
		rtsp.CSeqHeader:    strconv.Itoa(req.CSeq),
		rtsp.SessionHeader: req.SessionID,
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
		SessionID:  req.SessionID,
	}, nil
}

// handleTeardown handles TEARDOWN requests
func (h *RTSPHandler) handleTeardown(req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Verify session exists
	if req.SessionID == "" {
		return h.createErrorResponse(454, "Session Not Found", req.CSeq)
	}

	// Remove client from connection manager
	err := h.connectionManager.RemoveClient(req.SessionID)
	if err != nil {
		return h.createErrorResponse(500, "Internal Server Error", req.CSeq)
	}

	// Update stream state if no more clients
	stream := h.streamManager.GetStream(req.URL)
	if stream != nil {
		clients := h.connectionManager.GetClientsByStream(stream.ID)
		if len(clients) == 0 {
			stream.State = types.StreamInactive
			h.streamManager.UpdateStream(stream)
		}
	}

	// Prepare response headers
	headers := map[string]string{
		rtsp.CSeqHeader:    strconv.Itoa(req.CSeq),
		rtsp.SessionHeader: req.SessionID,
	}

	return &types.RTSPResponse{
		Version:    req.Version,
		StatusCode: 200,
		StatusText: rtsp.GetStatusCodeText(200),
		Headers:    headers,
	}, nil
}

// generateSDPForStream generates an SDP description for a stream
func (h *RTSPHandler) generateSDPForStream(stream *types.Stream) *sdp.SessionDescription {
	origin := sdp.Origin{
		Username:       "-",
		SessionID:      fmt.Sprintf("%d", time.Now().Unix()),
		SessionVersion: "1",
		NetType:        "IN",
		AddrType:       "IP4",
		UnicastAddr:    "127.0.0.1", // This should be the actual server address
	}

	sessionDesc := &sdp.SessionDescription{
		Version:     0,
		Origin:      origin,
		SessionName: stream.Name,
		SessionInfo: stream.Name + " Stream",
		Connection: &sdp.ConnectionData{
			NetType:        "IN",
			AddrType:       "IP4",
			ConnectionAddr: "127.0.0.1", // This should be the actual server address
		},
		Timing: &sdp.Timing{
			Start: 0, // 0 means session is not bounded by time
			Stop:  0,
		},
	}

	// Add media description based on stream format
	var mediaDesc *sdp.MediaDescription
	switch stream.MediaFormat {
	case types.H264MediaFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: "video 0 RTP/AVP 96", // 96 is dynamic payload type for H264
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "96 H264/90000"}, // 90000 Hz for video
				{Key: "fmtp", Value: "96 profile-level-id=42e01f; packetization-mode=1"},
				{Key: "control", Value: "track1"},
			},
		}
	case types.H265MediaFormat:
		mediaDesc = &sdp.MediaDescription{
			MediaName: "video 0 RTP/AVP 98", // 98 is dynamic payload type for H265
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "98 H265/90000"}, // 90000 Hz for video
				{Key: "fmtp", Value: "98 profile-space=0; profile-id=1; tier-flag=0; level-id=93"},
				{Key: "control", Value: "track1"},
			},
		}
	default:
		// Default to H264
		mediaDesc = &sdp.MediaDescription{
			MediaName: "video 0 RTP/AVP 96",
			Attributes: []sdp.Attribute{
				{Key: "rtpmap", Value: "96 H264/90000"},
				{Key: "fmtp", Value: "96 profile-level-id=42e01f; packetization-mode=1"},
				{Key: "control", Value: "track1"},
			},
		}
	}

	sessionDesc.MediaDesc = append(sessionDesc.MediaDesc, mediaDesc)

	return sessionDesc
}

// createErrorResponse creates an error response
func (h *RTSPHandler) createErrorResponse(statusCode int, statusText string, cseq int) (*types.RTSPResponse, error) {
	headers := map[string]string{
		rtsp.CSeqHeader: strconv.Itoa(cseq),
	}

	return &types.RTSPResponse{
		Version:    rtsp.RTSPVersion,
		StatusCode: statusCode,
		StatusText: statusText,
		Headers:    headers,
	}, nil
}
