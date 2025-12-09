package device_manage

import (
	"net/http"

	logger "yunyez/internal/pkg/logger"
	deviceService "yunyez/internal/service/device"

	"github.com/gin-gonic/gin"
)

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
	sn := c.Param("sn")
	if sn == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "sn is required",
			"Data": nil,
		})
		return
	}

	err := deviceService.ServiceInstance.DeleteDevice(c, sn)
	if err != nil {
		logger.Error(c.Request.Context(), "DeleteDevice failed", map[string]interface{}{
			"sn": sn,
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"Code": http.StatusInternalServerError,
			"Message": "DeleteDevice failed",
			"Data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Code": http.StatusOK,
		"Message": "DeleteDevice success",
		"Data": nil,
	})
}