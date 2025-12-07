package constant

// postgresql 相关常量

// 枚举类型
// 设备状态常量
const (
	// DeviceStatus 设备状态 -- 对应 device_global_status 枚举
	DeviceStatusDisabled    = "disabled"    // 已禁用
	DeviceStatusScrapped    = "scrapped"    // 已报废
	DeviceStatusActivated   = "activated"   // 已激活
	DeviceStatusInactivated = "inactivated" // 未激活

	// ConnectStatus 网络连接状态 -- 对应 connect_status_enum 枚举
	ConnectStatusConnected    = "connected"    // 已连接
	ConnectStatusDisconnected = "disconnected" // 未连接

	// PowerStatus 设备供电状态 -- 对应 power_status_enum 枚举
	PowerStatusPowerOn  = "power_on" // 已开启
	PowerStatusShutDown = "shutdown" // 已关闭
	PowerStatusStandby  = "standby"  // 待机
	PowerStatusCharging = "charging" // 充电中

	// WorkingStatus 工作状态 -- 对应 working_status_enum 枚举
	WorkingStatusIdle   = "idle"   // 空闲
	WorkingStatusActive = "active" // 活动中
	WorkingStatusBusy   = "busy"   // 忙碌
)


// vendor 厂商
const (
	Vendor_Public = "public"
	Vendor_Yunyez = "yunyez"
	Vendor_Test   = "test"
)

// GetVendor 获取厂商名称映射
func GetVendor() map[string]string {
	return map[string]string{
		Vendor_Public: "公版",
		Vendor_Yunyez: "云也子",
		Vendor_Test:   "测试",
	}
}
