package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yunyez/internal/video"
	"yunyez/internal/video/types"
)

func main() {
	// Create stream manager
	streamManager := video.NewStreamManager()

	// Create connection manager
	connectionManager := video.NewConnectionManager(streamManager)

	// Create RTSP server
	rtspServer := video.NewRTSPServer(":8554", streamManager, connectionManager)

	// Add a sample stream
	sampleStream := &types.Stream{
		ID:           "mystream",
		Name:         "Sample Stream",
		State:        types.StreamInactive,
		MediaType:    types.VideoMediaType,
		MediaFormat:  types.H264MediaFormat,
		Resolution:   "1920x1080",
		Bitrate:      2048,
		Framerate:    30.0,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Source:       "simulated",
	}
	
	err := rtspServer.AddStream(sampleStream)
	if err != nil {
		log.Printf("Failed to add sample stream: %v", err)
	} else {
		log.Printf("Added sample stream: %s", sampleStream.Name)
	}

	// Create context for the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the RTSP server
	fmt.Println("Starting RTSP server on :8554...")
	if err := rtspServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start RTSP server: %v", err)
	}

	// Print server status
	time.Sleep(1 * time.Second)
	status := rtspServer.GetStatus()
	fmt.Printf("RTSP Server Status:\n")
	fmt.Printf("- Uptime: %v\n", status.Uptime)
	fmt.Printf("- Total Streams: %d\n", status.TotalStreams)
	fmt.Printf("- Active Streams: %d\n", status.ActiveStreams)
	fmt.Printf("- Current Sessions: %d\n", status.CurrentSessions)

	// Wait for interrupt signal to stop the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down RTSP server...")
	
	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop the RTSP server
	if err := rtspServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping RTSP server: %v", err)
	} else {
		fmt.Println("RTSP server stopped successfully")
	}
}