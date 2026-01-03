package device

// device请求参数

import (
	"time"
	"yunyez/internal/common/constant"
	types "yunyez/internal/types/common"
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

// DeviceListRequest 获取设备列表请求参数
type DeviceListRequest struct {
	Page types.Page `form:",inline"`
	VendorName string `form:"vendorName,omitempty" default:""`
	DeviceType string `form:"deviceType,omitempty" default:""`
	Status int `form:"status,omitempty" oneof:"-1 1 2" default:"-1"`
}

// DeviceListResponse 获取设备列表响应参数
type DeviceListResponse struct {
	Page types.Page `json:"page,inline"`
	Total int `json:"total"`
	List []*DeviceListItem `json:"list"` // 设备列表元信息 - pure
}

// DeviceDetailResponse 获取设备详情响应参数
type DeviceDetailResponse struct {
	DeviceDetail DeviceDetail `json:"deviceDetail"` // 设备详情(加网络信息)
}

// DeviceBaseUpdateRequest 更新设备基本信息请求参数
type DeviceBaseUpdateRequest struct {
	SN string `json:"sn"` // 必填，作为 public ID

	// 以下字段均为可选，只有非 nil 才表示要更新 (区分未设置和设置为 nil 的情况)
	FirmwareVersion *string    `json:"firmwareVersion,omitempty"`
	Status          *string    `json:"status,omitempty" oneof:"activated inactivated disabled scrapped"`
	ExpireDate      *time.Time `json:"expireDate,omitempty"`
	ActivationTime  *time.Time `json:"activationTime,omitempty"`
	Remark          *string    `json:"remark,omitempty"`
}