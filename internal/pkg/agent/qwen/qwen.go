// Package qwen 提供 qwen 对话模型的实现
package qwen

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"yunyez/internal/common/config"
	"yunyez/internal/pkg/logger"
)

// qwen 配置
// 这里的package级别的变量只会在初始化时赋值一次，如果配置文件改变，需要重启服务才能生效
// 这种比较底层服务的配置一般不会频繁改变，所以放在package级别变量中是合理的
var (
	ChatModel = config.GetString("qwen.model")
	BaseURL   = config.GetString("qwen.endpoint")
	APIKey    = config.GetString("qwen.api_key")

	systemDesc = config.GetString("qwen.systemDesc")	
	role   = config.GetString("qwen.params.role")
	stream = config.GetBool("qwen.params.stream")

	timeout = 10 * time.Second
	size    = 10 // 响应通道缓冲区大小
)

// QwenChat 调用 Qwen 模型并返回响应通道（支持流式/非流式）
// 参数：
// - ctx context.Context: 上下文
// - message string: 对话内容
// 返回值：
// - <-chan string: 流式响应通道-只读（每个元素为一个片段）
// - error: 错误信息
func QwenChat(ctx context.Context, message string) (<-chan string, error) {
	var err error
	params, err := BuildQwenChatParams(ctx, message)
	if err != nil {
		logger.Error(ctx, "qwen.BuildQwenChatParams failed", map[string]any{
			"message": message,
			"error":   err.Error(),
		})
		return nil, err
	}
	resp, err := QwenChatHTTPRequest(ctx, params)
	if err != nil {
		logger.Error(ctx, "qwen.QwenChatHTTPRequest failed", map[string]any{
			"message": message,
			"error":   err.Error(),
		})
		return nil, err
	}

	out := make(chan string, size)
	go func() {
		defer close(out)        // 确保关闭通道
		defer resp.Body.Close() // 确保在函数退出时关闭响应体
		err = handleQwenChatResponse(ctx, resp, out)
		if err != nil {
			logger.Error(ctx, "qwen.handleQwenChatResponse failed", map[string]any{
				"message": message,
				"error":   err.Error(),
			})
			// 注意：不能往 out 发 error，只能 log
			// 调用方通过 context 或日志感知异常
		}
	}()
	return out, nil
}

// QwenChatHTTPRequest 调用 qwen 对话模型的 HTTP 请求
// 参数：
// - ctx context.Context: 上下文
// - param []byte: 请求参数
// 返回值：
// - *http.Response: HTTP 响应
// - error: 错误信息
func QwenChatHTTPRequest(ctx context.Context, param []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", BaseURL, bytes.NewBuffer(param))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call qwen service: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error(ctx, "Qwen API returned non-200 status", map[string]any{
			"status": resp.StatusCode,
			"body":   string(body),
		})
		return nil, fmt.Errorf("qwen API returned non-200 status: %d", resp.StatusCode)
	}

	return resp, nil
}

// BuildQwenChatParams 构建 qwen 对话模型的参数
// link：https://bailian.console.aliyun.com/?tab=api#/api/?type=model&url=2712576
// 参数：
// - ctx: 上下文
// - message string: 输入的对话消息
// - stream bool: 是否开启流式返回
// 返回值：
// - []byte: 构建后的参数
// - error: 错误信息
func BuildQwenChatParams(ctx context.Context, message string) ([]byte, error) {
	param := make(map[string]interface{})
	param["model"] = ChatModel
	param["messages"] = []map[string]string{
		{
			"role":    "system",
			"content": systemDesc,
		},
		{
			"role":    role,
			"content": message,
		},
	}
	param["stream"] = stream // 是否开启流式返回
	if stream {
		param["stream_options"] = map[string]interface{}{
			"include_usage": true,
		}
	}

	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	return paramBytes, nil
}

type QwenChatResponse struct {
	Output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"` // 对话输出内容
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct { // 模型使用统计
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// handleQwenChatResponse 处理 qwen 对话模型的 HTTP 响应
// 参数：
// - ctx context.Context: 上下文
// - resp *http.Response: HTTP 响应
// - res chan string: 回复通道-只写（每个元素为一个片段）
// 返回值：
// - string: 响应体字符串
// - error: 错误信息
func handleQwenChatResponse(ctx context.Context, resp *http.Response, res chan<- string) error {

	if !stream { // 非流式返回
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		var respBody QwenChatResponse
		err = json.Unmarshal(body, &respBody)
		if err != nil {
			return fmt.Errorf("unmarshal response body: %w", err)
		}
		if len(respBody.Output.Choices) <= 0 {
			logger.Error(ctx, "empty choices", map[string]any{
				"response": string(body),
			})
			return fmt.Errorf("empty choices")
		}
		res <- respBody.Output.Choices[0].Message.Content
		return nil
	}

	fmt.Printf("-------- handleQwenChatStreamResponse --------\n")

	err := handleQwenChatStreamResponse(ctx, resp, res)
	if err != nil {
		logger.Error(ctx, "handleQwenChatStreamResponse failed", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("handleQwenChatStreamResponse: %w", err)
	}

	return nil
}

// handleQwenChatStreamResponse 处理 qwen 对话模型的 HTTP 流式响应
// 参数：
// - ctx context.Context: 上下文
// - resp *http.Response: HTTP 响应
// - res chan string: 回复通道-只写（每个元素为一个片段）
// 返回值：
// - error: 错误信息
func handleQwenChatStreamResponse(ctx context.Context, resp *http.Response, res chan<- string) error {
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read response body: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" || line == ": ping" {
			continue
		}
		if line == "data: [DONE]" {
			break
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		jsonString := strings.TrimPrefix(line, "data: ")
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		err = json.Unmarshal([]byte(jsonString), &chunk)
		if err != nil {
			return fmt.Errorf("unmarshal response body: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content != "" {
			select {
			case res <- content: // 发送内容到通道
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		// end
		if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason == "stop" {
			break
		}
	}

	return nil
}
