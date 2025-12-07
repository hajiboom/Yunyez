package device

// device请求参数

import (
	"yunyez/internal/common/constant"
	types "yunyez/internal/types/common"

	deviceModel "yunyez/internal/model/device"
)

// GetDeviceStatus 获取设备状态
// 参数：
//  - status: 设备状态 这里-1表示全部，1表示在线，2表示离线【之所以不使用0是因为go的0值会与null产生歧义】
// 返回值：
//  - string: 设备状态常量
func GetDeviceStatus(status int) string {
	switch status {
	case -1: // 全部
		return ""
	case 1: // 激活
		return constant.DeviceStatusActivated
	case 2: // 未激活
		return constant.DeviceStatusInactivated
	case 3: // 禁用
		return constant.DeviceStatusDisabled
	case 4: // 报废
		return constant.DeviceStatusScrapped
	default:
		return ""
	}
}

type DeviceListRequest struct {
	Page types.Page `json:",inline"`
	VendorName string `json:"VendorName,omitempty" default:""`
	DeviceType string `json:"DeviceType,omitempty" default:""`
	Status int `json:"Status,omitempty" oneof:"-1 1 2" default:"-1"`
}

type DeviceListResponse struct {
	Page types.Page `json:",inline"`
	Total int `json:"Total"`
	Devices []*deviceModel.BaseDevice `json:"Devices"`
}