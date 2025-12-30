// Package mqtt MQTT连接管理
package mqtt

import (
	"context"
	"yunyez/internal/pkg/mqtt/core"
	"yunyez/internal/pkg/mqtt/middleware"

	logger "yunyez/internal/pkg/logger"

)

// ======================== MQTT连接 ======================================
// mqtt生命周期管理
// 业务处理回调

// StartConnect 启动MQTT连接
func StartConnect() error {
	// 初始化mqtt中间件
	middleware.InitMQTTMiddlewareChain()
	// 创建监听mqtt topic的客户端
	if err := core.InitMQTTClient(); err != nil {
		logger.Error(context.TODO(), "mqtt.client init error", map[string]interface{}{
			"error": err,
		})
		return err
	}
	return nil
}



