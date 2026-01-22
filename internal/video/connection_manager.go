// Package video implements the video stream management system
// It handles client connections, stream management, and RTSP protocol
package video

import (
	"fmt"
	"net"
	"sync"
	"time"

	"yunyez/internal/video/interfaces"
	"yunyez/internal/video/types"
)

// ConnectionManagerImpl implements the ConnectionManager interface
type ConnectionManagerImpl struct {
	clients map[string]*types.Client // sessionID -> Client
	streamClients map[string][]string // streamID -> []sessionIDs
	mutex   sync.RWMutex
	handler *RTSPHandler
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(streamMgr interfaces.StreamManager) *ConnectionManagerImpl {
	connMgr := &ConnectionManagerImpl{
		clients:     make(map[string]*types.Client),
		streamClients: make(map[string][]string),
	}

	// Create RTSP handler with the stream manager
	connMgr.handler = NewRTSPHandler(streamMgr, connMgr)

	return connMgr
}

// AddClient adds a new client connection
func (cm *ConnectionManagerImpl) AddClient(client *types.Client) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clients[client.SessionID]; exists {
		return fmt.Errorf("client with session ID %s already exists", client.SessionID)
	}

	cm.clients[client.SessionID] = client

	// Add to stream's client list
	cm.streamClients[client.StreamID] = append(cm.streamClients[client.StreamID], client.SessionID)

	return nil
}

// RemoveClient removes a client connection by session ID
func (cm *ConnectionManagerImpl) RemoveClient(sessionID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return fmt.Errorf("client with session ID %s does not exist", sessionID)
	}

	// Remove from stream's client list
	streamClients := cm.streamClients[client.StreamID]
	for i, sid := range streamClients {
		if sid == sessionID {
			// Remove element at index i
			cm.streamClients[client.StreamID] = append(streamClients[:i], streamClients[i+1:]...)
			break
		}
	}

	delete(cm.clients, sessionID)
	return nil
}

// GetClient retrieves a client by session ID
func (cm *ConnectionManagerImpl) GetClient(sessionID string) *types.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[sessionID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	clientCopy := *client
	return &clientCopy
}

// GetClientsByStream retrieves all clients connected to a specific stream
func (cm *ConnectionManagerImpl) GetClientsByStream(streamID string) []*types.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var clients []*types.Client
	sessionIDs, exists := cm.streamClients[streamID]
	if !exists {
		return clients
	}

	for _, sessionID := range sessionIDs {
		if client, exists := cm.clients[sessionID]; exists {
			// Return a copy to prevent external modification
			clientCopy := *client
			clients = append(clients, &clientCopy)
		}
	}

	return clients
}

// GetClients retrieves all connected clients
func (cm *ConnectionManagerImpl) GetClients() []*types.Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var clients []*types.Client
	for _, client := range cm.clients {
		// Return a copy to prevent external modification
		clientCopy := *client
		clients = append(clients, &clientCopy)
	}

	return clients
}

// UpdateClient updates client information
func (cm *ConnectionManagerImpl) UpdateClient(client *types.Client) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	existingClient, exists := cm.clients[client.SessionID]
	if !exists {
		return fmt.Errorf("client with session ID %s does not exist", client.SessionID)
	}

	// Update fields that can be changed
	existingClient.LastActivity = client.LastActivity
	existingClient.BytesTransferred = client.BytesTransferred

	return nil
}

// HandleRTSPRequest handles an incoming RTSP request
func (cm *ConnectionManagerImpl) HandleRTSPRequest(conn net.Conn, req *types.RTSPRequest) (*types.RTSPResponse, error) {
	// Update client activity if session exists
	if req.SessionID != "" {
		client := cm.GetClient(req.SessionID)
		if client != nil {
			client.LastActivity = time.Now()
			cm.UpdateClient(client)
		}
	}

	// Let the handler process the request
	return cm.handler.HandleRequest(conn, req)
}