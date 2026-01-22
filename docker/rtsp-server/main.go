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
		ID:          "cam1",
		Name:        "Camera 1",
		State:       types.StreamActive,
		MediaType:   types.VideoMediaType,
		MediaFormat: types.H264MediaFormat,
		Resolution:  "1920x1080",
		Bitrate:     2048,
		Framerate:   30.0,
		CreatedAt:   time.Now(),
		LastActivity: time.Now(),
		Source:      "rtsp://localhost:8554/cam1",
	}
	
	if err := streamManager.AddStream(sampleStream); err != nil {
		log.Fatalf("Failed to add sample stream: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the RTSP server
	fmt.Println("Starting RTSP server on :8554...")
	if err := rtspServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start RTSP server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	fmt.Println("RTSP server is running. Press Ctrl+C to stop.")
	<-sigChan

	fmt.Println("\nShutting down RTSP server...")
	
	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop the RTSP server
	if err := rtspServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping RTSP server: %v", err)
	}

	fmt.Println("RTSP server stopped.")
}