// Package tts provides text-to-speech services.
package tts

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "yunyez/internal/pkg/types/pb/ai"
)

// Protocol 协议类型
type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolGRPC  Protocol = "grpc"
)

// Service convert text to speech realtime interface
type Service interface {
	// Synthesize convert text to speech realtime
	Synthesize(ctx context.Context, text string) ([]byte, error)
	// Close 关闭客户端
	Close() error
}

// Config TTS 客户端配置
type Config struct {
	Model        string   // 模型类型：edge, chatTTS
	Protocol     Protocol // 协议类型：http, grpc
	HTTPEndpoint string   // HTTP 服务地址
	GRPCEndpoint string   // gRPC 服务地址
	// Edge TTS 参数
	Voice  string
	Rate   string
	Pitch  string
	Volume string
	// ChatTTS 参数
	Temperature string
}

// NewGRPCTTSClient 创建 gRPC TTS 客户端
func NewGRPCTTSClient(endpoint, voice string) (*GRPCTTSClient, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create gRPC connection: %w", err)
	}

	return &GRPCTTSClient{
		conn:   conn,
		client: pb.NewTTSServiceClient(conn),
		voice:  voice,
	}, nil
}

// GRPCTTSClient gRPC TTS 客户端
type GRPCTTSClient struct {
	conn   *grpc.ClientConn
	client pb.TTSServiceClient
	voice  string
}

// Close 关闭 gRPC 连接
func (c *GRPCTTSClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Synthesize 语音合成 - gRPC 方式
func (c *GRPCTTSClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	resp, err := c.client.Synthesize(ctx, &pb.SynthesizeRequest{
		Text:         text,
		Voice:        c.voice,
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       1.0,
		OutputFormat: "AUDIO_16KHZ_16BIT_RAW_PCM",
	})
	if err != nil {
		return nil, fmt.Errorf("call TTS service: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("TTS error: %s", resp.Error.Message)
	}

	return resp.AudioContent, nil
}

// NewTTSClient 创建 TTS 客户端（带配置参数）
func NewTTSClient(cfg Config) (Service, error) {
	switch cfg.Protocol {
	case ProtocolGRPC:
		return NewGRPCTTSClient(cfg.GRPCEndpoint, cfg.Voice)
	case ProtocolHTTP, "":
		return NewHTTPTTSClient(cfg), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %s", cfg.Protocol)
	}
}

// NewHTTPTTSClient 创建 HTTP TTS 客户端（向后兼容）
func NewHTTPTTSClient(cfg Config) Service {
	switch cfg.Model {
	case "chat", "chatTTS":
		return &ChatTTS{
			config: ChatTTSConfig{
				Endpoint:    cfg.HTTPEndpoint,
				Voice:       cfg.Voice,
				Rate:        cfg.Rate,
				Pitch:       cfg.Pitch,
				Volume:      cfg.Volume,
				Temperature: cfg.Temperature,
			},
		}
	case "edge", "edgeTTS", "":
		return &EdgeTTS{
			config: EdgeTTSConfig{
				Endpoint: cfg.HTTPEndpoint,
				Voice:    cfg.Voice,
				Rate:     cfg.Rate,
				Pitch:    cfg.Pitch,
				Volume:   cfg.Volume,
			},
			client: &http.Client{},
		}
	default:
		return &EdgeTTS{
			config: EdgeTTSConfig{
				Endpoint: cfg.HTTPEndpoint,
				Voice:    cfg.Voice,
				Rate:     cfg.Rate,
				Pitch:    cfg.Pitch,
				Volume:   cfg.Volume,
			},
			client: &http.Client{},
		}
	}
}
