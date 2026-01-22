package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"yunyez/internal/video"
	"yunyez/internal/video/types"
)

func main() {
	// 创建流管理器
	streamManager := video.NewStreamManager()

	// 创建连接管理器
	connectionManager := video.NewConnectionManager(streamManager)

	// 创建RTSP服务器
	rtspServer := video.NewRTSPServer(":8554", streamManager, connectionManager)

	// 添加示例流
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

	fmt.Println("Added sample stream:", sampleStream.Name)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动RTSP服务器
	fmt.Println("Starting RTSP server on :8554...")
	if err := rtspServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start RTSP server: %v", err)
	}

	// 显示服务器状态
	status := rtspServer.GetStatus()
	fmt.Printf("Server started. Active streams: %d, Current sessions: %d\n", 
		status.ActiveStreams, status.CurrentSessions)

	// 列出所有流
	streams := rtspServer.GetStreams()
	fmt.Printf("Available streams: %d\n", len(streams))
	for _, stream := range streams {
		fmt.Printf("  - %s (%s)\n", stream.Name, stream.State)
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 正常退出
	fmt.Println("Stopping RTSP server...")
	if err := rtspServer.Stop(ctx); err != nil {
		log.Printf("Error stopping RTSP server: %v", err)
	}

	fmt.Println("RTSP server stopped.")
}