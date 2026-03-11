// Package nlu NLU 意图识别客户端
package nlu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "yunyez/internal/pkg/types/pb/ai"
	pbCommon "yunyez/internal/pkg/types/pb/common"
)

// Protocol 协议类型
type Protocol string

const (
	ProtocolHTTP  Protocol = "http"
	ProtocolGRPC  Protocol = "grpc"
)

var (
	once sync.Once
	NLUClient Client
)

// Input 意图原始输入
type Input struct {
	Text string `json:"text"` // 输入的文本
}

// Intent NLU 意图识别结果
type Intent struct {
	Text       string  `json:"text"`       // 输入的文本
	Intent     string  `json:"intent"`     // 意图
	Confidence float32 `json:"confidence"` // 置信度
	IsCommand  bool    `json:"is_command"` // 是否为命令意图
}

// Emotion 情感识别结果
type Emotion struct {
	Text       string  `json:"text"`       // 输入的文本
	Emotion    string  `json:"emotion"`    // 情感
	Confidence float32 `json:"confidence"` // 置信度
}

// Config NLU 客户端配置
type Config struct {
	Model      string   // 模型类型：local
	Protocol   Protocol // 协议类型：http, grpc
	HTTPEndpoint string // HTTP 服务地址
	GRPCEndpoint string // gRPC 服务地址
}

// Client NLU 客户端
type Client struct {
	Protocol Protocol
	httpEndpoint string // NLU 服务地址
	grpcClient *GRPCClient
}

// GRPCClient gRPC NLU 客户端
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.NLUServiceClient
}

// NewClient 创建 NLU 客户端
func NewClient(cfg Config) Client {
	once.Do(func() {
		client := Client{
			Protocol:     cfg.Protocol,
			httpEndpoint: cfg.HTTPEndpoint,
		}
		
		if cfg.Protocol == ProtocolGRPC {
			grpcClient, err := newGRPCClient(cfg.GRPCEndpoint)
			if err != nil {
				panic(fmt.Sprintf("create gRPC client failed: %v", err))
			}
			client.grpcClient = grpcClient
		}
		
		NLUClient = client
	})
	return NLUClient
}

// newGRPCClient 创建 gRPC 客户端
func newGRPCClient(endpoint string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create gRPC connection: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewNLUServiceClient(conn),
	}, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	if c.grpcClient != nil && c.grpcClient.conn != nil {
		return c.grpcClient.conn.Close()
	}
	return nil
}

// Health 检查 NLU 服务健康状态
func (c *Client) Health() error {
	if c.Protocol == ProtocolGRPC {
		return c.healthGRPC()
	}
	return c.healthHTTP()
}

// healthHTTP HTTP 方式健康检查
func (c *Client) healthHTTP() error {
	httpReq, err := http.NewRequest("GET", c.httpEndpoint+"/health", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call NLU service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NLU service returned %d", resp.StatusCode)
	}

	return nil
}

// healthGRPC gRPC 方式健康检查
func (c *Client) healthGRPC() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.grpcClient.client.Health(ctx, &pbCommon.HealthRequest{})
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != "ok" {
		return fmt.Errorf("unhealthy status: %s", resp.Status)
	}

	return nil
}

// Predict 意图识别
func (c *Client) Predict(input *Input) (*Intent, error) {
	if c.Protocol == ProtocolGRPC {
		return c.predictGRPC(input)
	}
	return c.predictHTTP(input)
}

// predictHTTP HTTP 方式意图识别
func (c *Client) predictHTTP(input *Input) (*Intent, error) {
	reqBody := Input{Text: input.Text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.httpEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call NLU service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NLU service returned %d", resp.StatusCode)
	}

	var result Intent
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// predictGRPC gRPC 方式意图识别
func (c *Client) predictGRPC(input *Input) (*Intent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.client.Predict(ctx, &pb.PredictRequest{
		Text: input.Text,
	})
	if err != nil {
		return nil, fmt.Errorf("call NLU service: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("NLU error: %s", resp.Error.Message)
	}

	return &Intent{
		Text:       resp.Text,
		Intent:     resp.Intent,
		Confidence: resp.Confidence,
		IsCommand:  resp.IsCommand,
	}, nil
}

// EmotionJudge 情感识别
func (c *Client) EmotionJudge(text string) (*Emotion, error) {
	// TODO 文字情感识别
	return nil, nil
}
