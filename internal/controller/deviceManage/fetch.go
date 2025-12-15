package device_manage

import (
	"net/http"
	deviceService "yunyez/internal/service/device"
	deviceType "yunyez/internal/types/device"
	commonType "yunyez/internal/types/common"
	logger "yunyez/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// FetchDeviceList 获取设备列表
// @Summary 获取设备列表
// @Description 获取所有设备列表
// @Tags 设备管理
// @Accept json
// @Produce json
// @Success 200 {array} device.Device "成功获取设备列表"
// @Failure 500 {object} map[string]interface{} "获取设备列表失败"
// @Router /device/fetch [get]
func FetchDeviceList(c *gin.Context) {
	var req deviceType.DeviceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "Invalid request parameter",
			"Data": nil,
		})
		return
	}

	// 过滤条件
	filter := make(map[string]interface{})
	if req.VendorName != "" {
		filter["vendor_name"] = req.VendorName
	}
	if req.DeviceType != "" {
		filter["device_type"] = req.DeviceType
	}
	statusStr := deviceType.GetDeviceStatus(req.Status)
	if statusStr != "" {
		filter["status"] = statusStr
	}

	// 调用服务层获取设备列表
	devices, total, err := deviceService.ServiceInstance.ListDevices(c, req.Page.PageNum, req.Page.PageSize, filter)	
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to list devices", map[string]any{
			"error": err.Error(),
			"req": req,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"Code": http.StatusInternalServerError,
			"Message": "Failed to list devices",
			"Data": nil,
		})
		return
	}

	response := deviceType.DeviceListResponse{
		Page:     commonType.Page{         // ← 使用实际使用的分页值（可加兜底）
			PageNum:  req.Page.PageNum,
			PageSize: req.Page.PageSize,
		},
		Total:    int(total),
		Devices:  devices,
	}
	
	// 返回设备列表
	c.JSON(http.StatusOK, gin.H{
		"Code": http.StatusOK,
		"Message": "Success",
		"Data": response,
	})
}

// FetchDeviceDetail 获取设备详情
// @Summary 获取设备详情
// @Description 获取指定设备的详细信息
// @Tags 设备管理
// @Accept json
// @Produce json
// @Param sn path string true "设备序列号"
// @Success 200 {object} device.Device "成功获取设备详情"
// @Failure 400 {object} map[string]interface{} "无效的请求参数"
// @Failure 500 {object} map[string]interface{} "获取设备详情失败"
// @Router /device/detail/{sn} [get]
func FetchDeviceDetail(c *gin.Context) {
	sn := c.Param("sn")
	if sn == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"Code": http.StatusBadRequest,
			"Message": "Invalid request parameter",
			"Data": nil,
		})
		return
	}
	// 设备基本信息
	deviceBaseInfo, err := deviceService.ServiceInstance.GetDeviceBySN(c, sn)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to get device by SN", map[string]any{
			"error": err.Error(),
			"sn": sn,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 设备网络信息
	networkInfo, err := deviceService.ServiceInstance.GetDeviceNetworkBySN(c, sn)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to get device network info", map[string]any{
			"error": err.Error(),
			"sn": sn,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	deviceInfo := &deviceType.DeviceDetail{
		DeviceBaseInfo: deviceBaseInfo,
		DeviceNetworkInfo: networkInfo,
	}
	
	// 返回设备详情
	c.JSON(http.StatusOK, gin.H{
		"Code": http.StatusOK,
		"Message": "Success",
		"Data": deviceInfo,
	})
}