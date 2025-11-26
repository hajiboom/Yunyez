package constant

// postgresql 相关常量

// 枚举类型
// 设备状态常量
const (
	DeviceStatusActivated   = "activated"   // 已激活
	DeviceStatusInactivated = "inactivated" // 未激活

	ConnectStatusConnected    = "connected"    // 已连接
	ConnectStatusDisconnected = "disconnected" // 未连接

	PowerStatusOn  = "power_on"  // 已开启
	PowerStatusOff = "power_off" // 已关闭

	WorkingStatusIdle   = "idle"   // 空闲
	WorkingStatusActive = "active" // 活动中
	WorkingStatusBusy   = "busy"   // 忙碌
)
