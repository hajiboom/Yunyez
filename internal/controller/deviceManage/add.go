package device_manage

import "github.com/gin-gonic/gin"

// @Summary 添加设备
// @Description 添加新设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body device.Device true "设备信息"
// @Success 200 {object} map[string]interface{} "成功添加设备"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "添加设备失败"
// @Router /device/add [post]
func AddDevice(c *gin.Context) {
	
}
