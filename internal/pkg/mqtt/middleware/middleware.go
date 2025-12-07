package middleware

import (
	"context"
	"sync"
	"yunyez/internal/pkg/logger"

	paho "github.com/eclipse/paho.mqtt.golang"
)	


var (
	once sync.Once
	GlobalMQTTMiddlewareChain *MQTTMiddlewareChain // 全局MQTT中间件链
)

// MQTTMiddleware 定义MQTT中间件接口
type MQTTMiddleware interface {
	// Process 处理MQTT消息 返回处理后的context和是否继续处理
	Process(ctx context.Context, client paho.Client, msg paho.Message) (context.Context, bool)
}

// MQTTMiddlewareChain 定义MQTT中间件链
type MQTTMiddlewareChain struct {
	middlewareChain []MQTTMiddleware
}

// BuildMQTTMiddlewareChain 构建MQTT中间件链
func BuildMQTTMiddlewareChain() *MQTTMiddlewareChain {
	var middlewareChain []MQTTMiddleware
	once.Do(func(){
		middlewareChain = make([]MQTTMiddleware, 0)
	})
	return &MQTTMiddlewareChain{
		middlewareChain: middlewareChain,
	}
}

// Add 添加MQTT中间件到链中
func (c *MQTTMiddlewareChain)Add(middleware MQTTMiddleware){
	c.middlewareChain = append(c.middlewareChain, middleware)
}

// Process 处理MQTT消息 依次调用链中的中间件
func (c *MQTTMiddlewareChain)Process(ctx context.Context, client paho.Client, msg paho.Message) (context.Context, error){
	for _, middleware := range c.middlewareChain{
		ctx, ok := middleware.Process(ctx, client, msg)
		if !ok{
			logger.Info(ctx, "mqtt.middleware stop process", map[string]interface{}{
				"topic": msg.Topic(),
			})
			break
		}
	}
	return ctx, nil
}

// InitMQTTMiddlewareChain 初始化MQTT中间件链
func InitMQTTMiddlewareChain() {
	GlobalMQTTMiddlewareChain = BuildMQTTMiddlewareChain()
	// 添加设备中间件
	GlobalMQTTMiddlewareChain.Add(&DeviceMiddleware{})
	// 添加限流中间件
	GlobalMQTTMiddlewareChain.Add(&RateLimitMiddleware{})
	// 添加认证中间件
	GlobalMQTTMiddlewareChain.Add(&AuthMiddleware{})
}