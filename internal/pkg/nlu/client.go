package nlu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	config "yunyez/internal/common/config"
)

var (
	NRLAddress = config.GetString("nlu.address")
	once sync.Once
	NLUClient *Client
)

// Input 意图原始输入
type Input struct {
	Text string `json:"text"` // 输入的文本
}

// NLU 意图识别结果
type Intent struct {
	Text       string  `json:"text"`       // 输入的文本
	Intent     string  `json:"intent"`     // 意图
	Confidence float32 `json:"confidence"` // 置信度
	IsCommand  bool    `json:"is_command"` // 是否为命令意图
}

type Client struct {
	Address string `json:"address"` // NLU 服务地址
}

func NewClient() *Client {
	once.Do(func() {
		NLUClient = &Client{
			Address: NRLAddress,
		}
	})
	return NLUClient
}

// Health 检查 NLU 服务健康状态
// 在首次创建客户端时调用，检查 NLU 服务是否健康
func (c *Client) Health() error {
	httpReq, err := http.NewRequest("GET", c.Address+"/health", nil)
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

// Predict 意图识别
func (c *Client) Predict(input *Input) (*Intent, error) {
	reqBody := Input{Text: input.Text}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Address, bytes.NewBuffer(jsonData))
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
