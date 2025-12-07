package device

import (
	"context"
	"fmt"
	"yunyez/internal/common/config"
	logger "yunyez/internal/pkg/logger"
	mqtt "yunyez/internal/pkg/mqtt"
)

// 设备服务入口
func Start() {
	fmt.Println("device current environment: ", config.GetString("app.env"))
	// 启动mqtt连接
	if err := mqtt.StartConnect(); err != nil {
		logger.Error(context.TODO(), "mqtt.connect error", map[string]interface{}{
			"error": err,
		})
		return
	}
	// 启动http服务
	HTTPStart()

	select{} // 阻塞主goroutine
}