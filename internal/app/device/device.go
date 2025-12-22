// Package device 设备服务
package device

import (
	"context"
	"fmt"
	"yunyez/internal/common/config"
	device_manage "yunyez/internal/controller/deviceManage"
	voice_manage "yunyez/internal/controller/voiceManage"
	logger "yunyez/internal/pkg/logger"
	mqtt "yunyez/internal/pkg/mqtt"


	"github.com/gin-gonic/gin"
)

// Start 设备服务入口
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


// 设备服务--http接口

// HTTPStart 启动http服务 设置路由监听端口
func HTTPStart() {
	// 初始化gin路由
	r := gin.Default()

	// TODO 跨域中间件
	// ...

	// 设备路由
	api := r.Group("/api")


	deviceGroup := api.Group("/device")
	{
		deviceGroup.GET("/fetch", device_manage.FetchDeviceList)       // 获取设备列表
		deviceGroup.DELETE("/delete/:sn", device_manage.DeleteDevice)  // 删除设备
		deviceGroup.GET("/fetch/:sn", device_manage.FetchDeviceDetail) // 获取设备详情
		deviceGroup.PUT("/update", device_manage.UpdateDeviceInfo)     // 更新设备
	}

	// 语音路由
	r.POST("/voice", voice_manage.UploadVoice) // 发送语音

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
