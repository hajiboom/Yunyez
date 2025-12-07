package middleware

import (
	"context"
	logger "yunyez/internal/pkg/logger"
	"yunyez/internal/pkg/mqtt/core"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// DeviceMiddleware 设备中间件
// 在处理MQTT消息前 设备处理相关的中间件

type DeviceMiddleware struct {}

// Process 处理MQTT消息 设备相关的中间件
// 1.提取设备序列号
func (d *DeviceMiddleware)Process(ctx context.Context, client paho.Client, msg paho.Message) (context.Context, bool) {
	// 提取设备序列号
	topic, err := core.TopicParse(msg.Topic())
	if err != nil {
		logger.Error(ctx, "mqtt.topic parse error", map[string]interface{}{
			"topic": msg.Topic(),
			"error": err,
		})
		return ctx, false
	}
	ctx = context.WithValue(ctx, "deviceSN", topic.DeviceSN)
	return ctx, true
}

type RateLimitMiddleware struct {}

// Process 处理MQTT消息 限流相关的中间件
func (r *RateLimitMiddleware)Process(ctx context.Context, client paho.Client, msg paho.Message) (context.Context, bool) {
	// TODO： 限流逻辑
	// 1. 检查设备是否超过限流阈值
	// 2. 如果超过阈值 则拒绝处理该消息
	// 3. 如果未超过阈值 则继续处理该消息
	
	return ctx, true
}

type AuthMiddleware struct {}

// Process 处理MQTT消息 认证相关的中间件
func (a *AuthMiddleware)Process(ctx context.Context, client paho.Client, msg paho.Message) (context.Context, bool) {
	// TODO： 认证逻辑
	// 1. 检查设备是否存在
	// 2. 如果不存在 则拒绝处理该消息
	// 3. 如果存在 则继续处理该消息
	
	return ctx, true
}


