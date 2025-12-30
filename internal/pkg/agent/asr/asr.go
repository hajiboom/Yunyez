// Package asr 语音识别服务
package asr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)


// Service 语音识别服务接口
type Service interface {
	// Transfer 语音识别 音频转换为文本
	Transfer(ctx context.Context, data []byte) (string, error)
}

func NewASRClient(model, endpoint string) Service {
	// 选择 ASR 模型
	switch model {
	case "local":
		return &LocalASRClient{
			Endpoint: endpoint,
			client:   &http.Client{},
		}
	default:
		panic(fmt.Sprintf("unknown ASR model: %s", model))
	}
}

type LocalASRClient struct {
	Endpoint string `json:"endpoint"` // ASR 服务地址
	client   *http.Client // HTTP 客户端
}

// Transfer 语音识别 音频转换为文本 -- 本地模型
// 参数：
//   - ctx: 上下文对象，用于取消操作和传递上下文信息
//   - data: 包含音频数据的字节切片
// 返回值：
//   - string: 转换后的文本结果
//   - error: 操作过程中遇到的错误
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