package device_manage

import "github.com/gin-gonic/gin"

// @Summary 删除设备
// @Description 删除指定设备
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param sn path string true "设备序列号"
// @Success 200 {object} map[string]interface{} "成功删除设备"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "删除设备失败"
// @Router /device/delete/{sn} [delete]
func DeleteDevice(c *gin.Context) {
	
}