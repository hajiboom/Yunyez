package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	config "yunyez/internal/common/config"
	logger "yunyez/internal/pkg/logger"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// MQTT消息 转发结构体
type Message struct {
	Topic       string `json:"topic"`       // topic
	CommandType string `json:"commandType"` // 命令类型
	ClientID    string `json:"clientID"`    // 客户端ID-设备序列号
	Content     []byte `json:"content"`     // 消息体
	StartTime   string `json:"startTime"`   // 消息发送时间

}

// MQTT 消息处理策略
type Strategy interface {
	Send(ctx context.Context, msg *Message)
}

// -----------------------------------------------------------
// SendHandler 消息转发处理结构体
type SendHandler struct {
	Sender Strategy
}

// Set 设置消息转发策略
// 支持 HTTP / Kafka / 其他自定义协议
func (s *SendHandler) Set(model string, client *paho.Client) {
	switch model {
		case "http":
			s.Sender = &HttpStrategy{client: client}
		default:
			s.Sender = &KafkaStrategy{client: client}
	}
}

// Send  转发消息 - 消息处理策略
// 参数：
//   - ctx: 上下文
//   - msg: MQTT 消息
func (s *SendHandler) Send(ctx context.Context, msg *Message) {
	s.Sender.Send(ctx, msg)
}

// -----------------------------------------------------------

// HttpStrategy HTTP 消息处理策略
type HttpStrategy struct{
	client *paho.Client // MQTT 客户端实例
}

// Send  转发消息 - HTTP 消息处理策略
// 参数：
//   - ctx: 上下文
//   - msg: MQTT 消息
func (s *HttpStrategy) Send(ctx context.Context, msg *Message) {

	httpAddr := config.GetString("http.addr")
	httpPort := config.GetString("http.port")
		
	// topic 对应的 HTTP 接口地址
	url := fmt.Sprintf("%s:%s/%s", httpAddr, httpPort, msg.CommandType)
	// 构造 HTTP 请求
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Topic", msg.Topic)
	header.Set("ClientID", msg.ClientID)
	header.Set("StartTime", msg.StartTime)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(msg.Content))
	if err != nil {
		logger.Error(ctx, "Failed to create HTTP request", map[string]any{
			"error": err.Error(),
			"url": url,
		})
		return
	}
	req.Header = header
	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(ctx, "Failed to send HTTP request", map[string]any{
			"error": err.Error(),
			"url": url,
		})
		return
	}
	defer resp.Body.Close()
}

type KafkaStrategy struct{
	client *paho.Client // MQTT 客户端实例
}

// Send  转发消息 - Kafka 消息处理策略
// 参数：
//   - ctx: 上下文
//   - msg: MQTT 消息
func (s *KafkaStrategy) Send(ctx context.Context, msg *Message) {

}