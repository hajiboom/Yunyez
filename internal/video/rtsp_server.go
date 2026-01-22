package video

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"yunyez/internal/pkg/rtsp"
	"yunyez/internal/pkg/transport"
	"yunyez/internal/video/interfaces"
	"yunyez/internal/video/types"
)

// RTSPServer represents the main RTSP server
type RTSPServer struct {
	address         string
	transport       *transport.TCPTransport
	streamManager   interfaces.StreamManager
	connectionManager interfaces.ConnectionManager
	httpServer      *http.Server
	stats           *types.ServerStats
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	startTime       time.Time
}

// NewRTSPServer creates a new RTSP server instance
func NewRTSPServer(address string, streamMgr interfaces.StreamManager, connMgr interfaces.ConnectionManager) *RTSPServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	server := &RTSPServer{
		address:         address,
		streamManager:   streamMgr,
		connectionManager: connMgr,
		ctx:             ctx,
		cancel:          cancel,
		stats: &types.ServerStats{
			StartTime: time.Now(),
		},
	}
	
	// Initialize HTTP server for health checks
	server.httpServer = &http.Server{
		Addr:    ":8080", // Default HTTP port for health checks
		Handler: server.createHTTPHandler(),
	}
	
	return server
}

// Start starts the RTSP server
func (s *RTSPServer) Start(ctx context.Context) error {
	s.mutex.Lock()
	s.startTime = time.Now()
	s.mutex.Unlock()

	// Initialize TCP transport
	s.transport = transport.NewTCPTransport()
	
	// Start listening on the RTSP port
	err := s.transport.Listen(s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.address, err)
	}

	log.Printf("RTSP server listening on %s", s.address)

	// Start accepting connections in a goroutine
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := s.transport.Accept()
				if err != nil {
					log.Printf("Failed to accept connection: %v", err)
					continue
				}

				// Handle the connection in a separate goroutine
				go s.handleConnection(conn)
			}
		}
	}()

	// Start HTTP server for health checks in a goroutine
	go func() {
		log.Printf("Starting HTTP server for health checks on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the RTSP server
func (s *RTSPServer) Stop(ctx context.Context) error {
	s.cancel()

	// Close RTSP transport
	if s.transport != nil {
		if err := s.transport.Close(); err != nil {
			log.Printf("Error closing RTSP transport: %v", err)
		}
	}

	// Shutdown HTTP server gracefully
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}

	log.Println("RTSP server stopped")
	return nil
}

// handleConnection handles an incoming RTSP connection
func (s *RTSPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	// Process requests from the connection
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Set a read deadline to prevent hanging
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			// Parse the RTSP request
			req, err := rtsp.ParseRequestFromBytes(readRequest(conn))
			if err != nil {
				log.Printf("Failed to parse RTSP request: %v", err)
				return
			}

			// Convert to internal request type
			internalReq := &types.RTSPRequest{
				Method:    req.Method,
				URL:       req.URL.String(),
				Version:   req.Version,
				Headers:   req.Headers,
				Body:      req.Body,
				CSeq:      req.CSeq,
				SessionID: req.SessionID,
			}

			// Handle the request using the connection manager
			resp, err := s.connectionManager.HandleRTSPRequest(conn, internalReq)
			if err != nil {
				log.Printf("Failed to handle RTSP request: %v", err)
				return
			}

			// Send the response back to the client
			_, err = conn.Write(respToBytes(resp))
			if err != nil {
				log.Printf("Failed to send response: %v", err)
				return
			}

			// Update stats
			s.updateStatsAfterRequest(req, resp)
		}
	}
}

// readRequest reads an RTSP request from the connection
func readRequest(conn net.Conn) []byte {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil
	}
	return buf[:n]
}

// respToBytes converts an internal response to bytes
func respToBytes(resp *types.RTSPResponse) []byte {
	rtspResp := rtsp.NewResponse(resp.StatusCode, resp.StatusText, resp.Headers, resp.Body)
	if resp.SessionID != "" {
		rtspResp.AddHeader(rtsp.SessionHeader, resp.SessionID)
	}
	if resp.RTPInfo != nil {
		rtpInfo := fmt.Sprintf("url=%s;seq=%d;rtptime=%d", resp.RTPInfo.URL, resp.RTPInfo.Seq, resp.RTPInfo.RTPTime)
		rtspResp.AddHeader("RTP-Info", rtpInfo)
	}
	return rtspResp.Bytes()
}

// updateStatsAfterRequest updates server statistics after processing a request
func (s *RTSPServer) updateStatsAfterRequest(req *rtsp.Request, resp *types.RTSPResponse) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Update stats based on request/response
	if req.Body != nil {
		s.stats.BytesReceived += int64(len(req.Body))
	}
	if resp.Body != nil {
		s.stats.BytesSent += int64(len(resp.Body))
	}

	// Update session counts based on method
	switch req.Method {
	case rtsp.Setup:
		s.stats.CurrentSessions++
		s.stats.TotalSessions++
	case rtsp.Teardown:
		if s.stats.CurrentSessions > 0 {
			s.stats.CurrentSessions--
		}
	}
}

// GetStatus returns the current status of the server
func (s *RTSPServer) GetStatus() *types.ServerStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Calculate uptime
	uptime := time.Since(s.startTime)

	// Get stream count
	streams := s.streamManager.GetStreams()
	activeStreams := 0
	for _, stream := range streams {
		if stream.State == types.StreamActive {
			activeStreams++
		}
	}

	// Get client count
	_ = s.connectionManager.GetClients() // Use the variable to avoid "declared but not used" error

	return &types.ServerStats{
		Uptime:          uptime,
		TotalSessions:   s.stats.TotalSessions,
		CurrentSessions: s.stats.CurrentSessions,
		TotalStreams:    len(streams),
		ActiveStreams:   activeStreams,
		BytesSent:       s.stats.BytesSent,
		BytesReceived:   s.stats.BytesReceived,
		StartTime:       s.stats.StartTime,
	}
}

// GetStream returns information about a specific stream
func (s *RTSPServer) GetStream(streamID string) *types.Stream {
	return s.streamManager.GetStream(streamID)
}

// GetStreams returns all active streams
func (s *RTSPServer) GetStreams() []*types.Stream {
	return s.streamManager.GetStreams()
}

// GetClient returns information about a specific client
func (s *RTSPServer) GetClient(sessionID string) *types.Client {
	return s.connectionManager.GetClient(sessionID)
}

// GetClients returns all connected clients
func (s *RTSPServer) GetClients() []*types.Client {
	return s.connectionManager.GetClients()
}

// AddStream adds a new stream to the server
func (s *RTSPServer) AddStream(stream *types.Stream) error {
	return s.streamManager.AddStream(stream)
}

// RemoveStream removes a stream from the server
func (s *RTSPServer) RemoveStream(streamID string) error {
	return s.streamManager.RemoveStream(streamID)
}

// GetHTTPHandler returns the HTTP handler for health checks and status
func (s *RTSPServer) GetHTTPHandler() http.Handler {
	return s.createHTTPHandler()
}

// createHTTPHandler creates the HTTP handler for health checks and status
func (s *RTSPServer) createHTTPHandler() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/status", s.statusHandler)
	mux.HandleFunc("/streams", s.streamsHandler)
	mux.HandleFunc("/clients", s.clientsHandler)
	
	return mux
}

// healthHandler handles health check requests
func (s *RTSPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	status := s.GetStatus()
	
	response := types.APIResponse{
		Status: "healthy",
		Data: map[string]interface{}{
			"uptime":       status.Uptime.String(),
			"start_time":   status.StartTime.Format(time.RFC3339),
			"total_streams": status.TotalStreams,
			"active_streams": status.ActiveStreams,
			"current_sessions": status.CurrentSessions,
			"bytes_sent":     status.BytesSent,
			"bytes_received": status.BytesReceived,
		},
	}
	
	writeJSONResponse(w, response, http.StatusOK)
}

// statusHandler handles status requests
func (s *RTSPServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := s.GetStatus()
	
	response := types.APIResponse{
		Status: "success",
		Data:   status,
	}
	
	writeJSONResponse(w, response, http.StatusOK)
}

// streamsHandler handles stream list requests
func (s *RTSPServer) streamsHandler(w http.ResponseWriter, r *http.Request) {
	streams := s.GetStreams()
	
	response := types.APIResponse{
		Status: "success",
		Data:   streams,
	}
	
	writeJSONResponse(w, response, http.StatusOK)
}

// clientsHandler handles client list requests
func (s *RTSPServer) clientsHandler(w http.ResponseWriter, r *http.Request) {
	clients := s.GetClients()
	
	response := types.APIResponse{
		Status: "success",
		Data:   clients,
	}
	
	writeJSONResponse(w, response, http.StatusOK)
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}