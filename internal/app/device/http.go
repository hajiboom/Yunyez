package device

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"yunyez/internal/common/config"
	device_manage "yunyez/internal/controller/deviceManage"
)

// 设备服务--http接口

// HTTPStart 启动http服务 设置路由监听端口
func HTTPStart() {
	// 初始化gin路由
	r := gin.Default()

	// TODO 跨域中间件
	// ...

	// 设备路由
	deviceGroup := r.Group("/device")
	{
		deviceGroup.POST("/add", device_manage.AddDevice)                // 添加设备
		deviceGroup.GET("/fetch", device_manage.FetchDeviceList)         // 获取设备列表
		deviceGroup.DELETE("/delete/:sn", device_manage.DeleteDevice)   // 删除设备
		deviceGroup.GET("/fetch/:sn", device_manage.FetchDeviceDetail) // 获取设备详情
		deviceGroup.PUT("/update/:sn", device_manage.UpdateDeviceInfo) // 更新设备
	}

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
