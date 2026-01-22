// Package constant 常量定义 - 默认
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
	VendorPublic = "public"
	VendorYunyez = "yunyez"
	VendorTest   = "test"
)



// GetVendor 获取厂商名称映射
func GetVendor() map[string]string {
	return map[string]string{
		VendorPublic: "公版",
		VendorYunyez: "云也子",
		VendorTest:   "测试",
	}
}

// intent 意图
// 【目前只是测试阶段，后续会根据实际场景进行调整】
// 可能会放弃意图分类采用无监督学习
const (
	IntentPlayMusic      = "play_music"      // 播放音乐
	IntentSetTemperature = "set_temperature" // 设置温度
	IntentTurnOnLight    = "turn_on_light"   // 开灯
	IntentTurnOffLight   = "turn_off_light"  // 关灯
	IntentChitChat       = "chit_chat"       // 闲聊
	IntentDenyAction     = "deny_action"     // 取消执行
)

// model 模型
const (
	ModelQwenLLM  = "qwen"  // 通义千问
	ModelLocalLLM = "local" // 本地模型

	ModelChatTTS = "chat" // ChatTTS
	ModelEdgeTTS = "edge" // EdgeTTS
)

