package device_manage

import "github.com/gin-gonic/gin"

// @Summary 更新设备信息
// @Description 更新指定设备的信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body device.Device true "设备信息"
// @Success 200 {object} map[string]interface{} "成功更新设备信息"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "更新设备信息失败"
// @Router /device/update [put]
func UpdateDeviceInfo(c *gin.Context) {
	
}