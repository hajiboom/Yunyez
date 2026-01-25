// Package device 设备服务
package device

import (
	"context"
	"fmt"
	"yunyez/internal/common/config"
	middleware "yunyez/internal/middleware"
	deviceManage "yunyez/internal/controller/deviceManage"
	voiceManage "yunyez/internal/controller/voiceManage"
	logger "yunyez/internal/pkg/logger"
	mqtt "yunyez/internal/pkg/mqtt"

	"github.com/gin-gonic/gin"
)

// Start 设备服务入口
func Start() {
	fmt.Println("device current environment: ", config.GetString("app.env"))

	// load config
	if err := config.Init(); err != nil {
		fmt.Printf("failed to init config: %v\n", err)
		return
	}

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


// 设备服务--http接口

// HTTPStart 启动http服务 设置路由监听端口
func HTTPStart() {
	// 初始化gin路由
	r := gin.New()

	// 添加中间件集合
	middlewares := middleware.SetupMiddlewares()
	for _, m := range middlewares {
		r.Use(m)
	}

	// 设备路由
	api := r.Group("/api")


	deviceGroup := api.Group("/device")
	{
		deviceGroup.GET("/fetch", deviceManage.FetchDeviceList)       // 获取设备列表
		deviceGroup.DELETE("/delete/:sn", deviceManage.DeleteDevice)  // 删除设备
		deviceGroup.GET("/fetch/:sn", deviceManage.FetchDeviceDetail) // 获取设备详情
		deviceGroup.PUT("/update", deviceManage.UpdateDeviceInfo)     // 更新设备
	}

	// 语音路由
	r.POST("/voice", voiceManage.UploadVoice) // 发送语音

	// 获取 HTTP 端口号
	port := ":" + config.GetString("http.port")
	// 启动 HTTP 服务
	err := r.Run(port)
	if err != nil {
		fmt.Printf("start http server on port %s failed: %v", port, err)
		return
	}
	fmt.Printf("start http server on port %s success", port)
}
