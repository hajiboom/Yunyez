// Package core provides the core functionality for MQTT client
// MQTTClient is a wrapper for MQTT client
package core

import (
	"context"
	"errors"
	"time"
	logger "yunyez/internal/pkg/logger"
	"yunyez/internal/pkg/mqtt/protocol/voice"

	paho "github.com/eclipse/paho.mqtt.golang"
)

var (
	ErrMQTTClientNotInit = errors.New("mqtt.client not init")
)

// Client 封装的一个自定义的 MQTT 客户端
type Client struct {
	Ctx       context.Context    // 上下文
	Cancel    context.CancelFunc // 取消函数
	Topic     Topic              // 订阅主题
	Client    paho.Client        // MQTT 客户端
	Qos       byte               // QoS 等级 0: 最多一次 / 1: 至少一次 / 2: 只有一次
	RequestID string             // 请求ID
	ReplayID  string             // 回复ID 防止对话回答错位
	Data      []byte             // 消息数据
}

// Init 初始化 MQTT 客户端
// 返回一个声明的mqtt客户端但是需要手动设置属性
func Init(ctx context.Context, topic Topic) *Client {
	var client Client
	ctx, cancel := context.WithCancel(ctx)
	client = Client{
		Ctx:       ctx,
		Cancel:    cancel,
		Topic:     topic,
		RequestID: topic.DeviceSN, // 请求ID 设备序列号
	}

	return &client
}

// GetMQTTClient 获取MQTT客户端结构体
// 参数:
//   - ctx: 上下文
//   - topic: 主题
// 返回值:
//   - Client: 自定义的 MQTT 客户端
//   - error: 错误信息
func GetMQTTClient(ctx context.Context, topic Topic) (Client, error) {
	client := Init(ctx, topic)
	client.Client = MqttClient
	return *client, nil
}

// Publish 发送消息到指定topic
// 参数:
//   - ctx: 上下文
//   - data: 消息数据 字节切片
//   - audioConfig: 语音配置[采样率、格式、声道数]
//
// 返回值:
//   - error: 错误信息
func (c *Client) Publish(ctx context.Context, data []byte, audioConfig voice.AudioConfig) error {
	if c.Client == nil {
		logger.Error(ctx, "mqtt.client not init", map[string]interface{}{
			"topic": c.Topic.String(),
			"error": ErrMQTTClientNotInit,
		})
		return ErrMQTTClientNotInit
	}
	payload := voice.BuildFullPayload(0, data, audioConfig)
	err := send(ctx, c.Client, c.Topic, c.Qos, payload)
	if err != nil {
		logger.Error(ctx, "mqtt.publishStream error", map[string]interface{}{
			"topic": c.Topic.String(),
			"error": err,
		})
		return err
	}

	logger.Info(ctx, "mqtt.publishStream success", map[string]interface{}{
		"topic":       c.Topic.String(),
		"format":      audioConfig.AudioFormat,
		"sample_rate": audioConfig.AudioSampleRate,
		"channel":     audioConfig.AudioChannel,
		"payload_len": len(payload),
	})

	return nil
}

// PublishStream 发送流式消息到指定topic
// 参数:
//   - ctx: 上下文
//   - seq: 消息序号 从0开始递增
//   - data: 消息数据 字节切片
//   - audioConfig: 语音配置[采样率、格式、声道数]
//   - isLast: 是否为最后一帧
//
// 返回值:
//   - error: 错误信息
func (c Client) PublishStream(ctx context.Context, seq uint16, data []byte, audioConfig voice.AudioConfig, isLast bool) error {
	if c.Client == nil {
		logger.Error(ctx, "mqtt.client not init", map[string]interface{}{
			"topic": c.Topic.String(),
			"error": ErrMQTTClientNotInit,
		})
		return ErrMQTTClientNotInit
	}
	payload := voice.BuildStreamPayload(seq, data, audioConfig, isLast)
	err := send(ctx, c.Client, c.Topic, c.Qos, payload)
	if err != nil {
		logger.Error(ctx, "mqtt.send error", map[string]interface{}{
			"topic": c.Topic.String(),
			"error": err,
		})
		return err
	}

	logger.Info(ctx, "mqtt.send success", map[string]interface{}{
		"topic":       c.Topic.String(),
		"seq":         seq,
		"isLast":      isLast,
		"format":      audioConfig.AudioFormat,
		"sample_rate": audioConfig.AudioSampleRate,
		"channel":     audioConfig.AudioChannel,
		"payload_len": len(payload),
	})

	return nil
}

// send 实际发送函数
// 内部调用mqtt客户端的Publish方法
func send(ctx context.Context, client paho.Client, topic Topic, qos byte, data []byte) error {

	topicStr := topic.String()
	token := client.Publish(topicStr, qos, false, data)
	token.Wait()
	select {
	case <-token.Done(): // 等待发布完成
		return token.Error()
	case <-ctx.Done(): // 上下文取消
		return ctx.Err()
	case <-time.After(5 * time.Second): // 超时
		return errors.New("mqtt.send timeout")
	default:
		return nil
	}
}
