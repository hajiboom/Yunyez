// Package asr 语音识别服务
package asr

import (
	"bytes"
	"context"
	"encoding/json"
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

// Service 语音识别服务接口
type Service interface {
	// Transfer 语音识别 音频转换为文本
	Transfer(ctx context.Context, data []byte) (string, error)
	// Close 关闭客户端
	Close() error
}

// Config ASR 客户端配置
type Config struct {
	Model      string   // 模型类型：local
	Protocol   Protocol // 协议类型：http, grpc
	HTTPEndpoint string // HTTP 服务地址
	GRPCEndpoint string // gRPC 服务地址
}

// NewASRClient 创建 ASR 客户端
func NewASRClient(cfg Config) (Service, error) {
	switch cfg.Protocol {
	case ProtocolGRPC:
		return NewGRPCClient(cfg.GRPCEndpoint)
	case ProtocolHTTP, "":
		return NewHTTPClient(cfg.HTTPEndpoint, cfg.Model)
	default:
		return nil, fmt.Errorf("unknown protocol: %s", cfg.Protocol)
	}
}

// NewHTTPClient 创建 HTTP 客户端（向后兼容）
func NewHTTPClient(endpoint, model string) (Service, error) {
	// 选择 ASR 模型
	switch model {
	case "local":
		return &LocalASRClient{
			Endpoint: endpoint,
			client:   &http.Client{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown ASR model: %s", model)
	}
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient(endpoint string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create gRPC connection: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewASRServiceClient(conn),
	}, nil
}

// LocalASRClient HTTP ASR 客户端
type LocalASRClient struct {
	Endpoint string
	client   *http.Client
}

// Transfer 语音识别 - HTTP 方式
func (c *LocalASRClient) Transfer(ctx context.Context, data []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create ASR request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call ASR service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ASR service returned %d", resp.StatusCode)
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode ASR response: %w", err)
	}

	return result.Text, nil
}

// Close 关闭 HTTP 客户端（无操作）
func (c *LocalASRClient) Close() error {
	return nil
}

// GRPCClient gRPC ASR 客户端
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.ASRServiceClient
}

// Close 关闭 gRPC 连接
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Transfer 语音识别 - gRPC 方式
func (c *GRPCClient) Transfer(ctx context.Context, data []byte) (string, error) {
	resp, err := c.client.Recognize(ctx, &pb.RecognizeRequest{
		AudioContent: data,
		Encoding:     "LINEAR16_PCM",
		SampleRate:   16000,
		NumChannels:  1,
		LanguageCode: "zh-CN",
	})
	if err != nil {
		return "", fmt.Errorf("call ASR service: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("ASR error: %s", resp.Error.Message)
	}

	return resp.Transcript, nil
}
