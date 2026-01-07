package device_manage

import (
	"net/http"

	deviceService "yunyez/internal/service/device"
	deviceType "yunyez/internal/types/device"
	logger "yunyez/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// UpdateDeviceInfo 更新设备信息
// @Summary 更新设备信息
// @Description 更新指定设备的信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param device body deviceType.DeviceBaseUpdateRequest true "设备基本信息"
// @Success 200 {object} map[string]interface{} "成功更新设备信息"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "更新设备信息失败"
// @Router /device/update [put]
func UpdateDeviceInfo(c *gin.Context) {
	var req deviceType.DeviceBaseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "invalid request body",
			"Data":  err.Error(),
		})
		return
	}
	// 校验字段
	if req.SN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "deviceSn is required",
			"Data": nil,
		})
		return
	}
	// 查询设备是否存在
	exist, err := deviceService.ServiceInstance.CheckDeviceExist(c.Request.Context(), req.SN)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to check device exist", map[string]any{
			"error": err.Error(),
			"req": req,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"Code": http.StatusInternalServerError,
			"Message": "check device exist failed",
			"Data": nil,
		})
		return
	}
	if !exist {
		logger.Info(c.Request.Context(), "Device not found", map[string]any{
			"req": req,
		})
		c.JSON(http.StatusNotFound, gin.H{
			"Code": http.StatusNotFound,
			"Message": "device not found",
			"Data": nil,
		})
		return
	}
	// 更新设备信息
	updates := make(map[string]interface{})
	if req.FirmwareVersion != nil {
		updates["firmware_version"] = *req.FirmwareVersion
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ExpireDate != nil {
		updates["expire_date"] = *req.ExpireDate
	}
	if req.ActivationTime != nil {
		updates["activation_time"] = *req.ActivationTime
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if len(updates) == 0 {
		logger.Error(c.Request.Context(), "No update fields provided", map[string]any{
			"req": req,
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "no update fields provided",
			"Data": nil,
		})
		return
	}
	// 更新设备信息
	err = deviceService.ServiceInstance.UpdateDevice(c.Request.Context(), req.SN, updates)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to update device info", map[string]any{
			"error": err.Error(),
			"req": req,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"Code": http.StatusInternalServerError,
			"Message": "update device info failed",
			"Data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Code": http.StatusOK,
		"Message": "update device info success",
	})
}