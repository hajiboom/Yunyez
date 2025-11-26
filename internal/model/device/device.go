package device

import (
	gorm "gorm.io/gorm"
	"time"
)

// BaseDevice 设备基础信息模型
type BaseDevice struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceSN        string         `gorm:"column:device_sn;type:varchar(64);not null;uniqueIndex" json:"device_sn"`            // 设备序列号
	IMEI            string         `gorm:"column:imei;type:varchar(32);uniqueIndex" json:"imei,omitempty"`                     // IMEI
	ICCID           string         `gorm:"column:iccid;type:varchar(32);uniqueIndex" json:"iccid,omitempty"`                   // ICCID
	DeviceType      string         `gorm:"column:device_type;type:varchar(32);not null" json:"device_type"`                    // 设备类型
	VendorID        int64          `gorm:"column:vendor_id;not null" json:"vendor_id"`                                         // 厂商ID
	VendorName      string         `gorm:"column:vendor_name;type:varchar(64);not null" json:"vendor_name"`                    // 厂商名称
	HardwareVersion string         `gorm:"column:hardware_version;type:varchar(32);not null" json:"hardware_version"`          // 硬件版本
	FirmwareVersion string         `gorm:"column:firmware_version;type:varchar(32);not null" json:"firmware_version"`          // 固件版本
	ProductModel    string         `gorm:"column:product_model;type:varchar(64);not null" json:"product_model"`                // 产品型号
	ManufactureDate time.Time      `gorm:"column:manufacture_date;not null" json:"manufacture_date"`                           // 生产日期
	ExpireDate      time.Time      `gorm:"column:expire_date" json:"expire_date,omitempty"`                                    // 质保到期日
	Status          string         `gorm:"column:status;type:device_global_status;not null;default:inactivated" json:"status"` // 设备状态
	ActivationTime  time.Time      `gorm:"column:activation_time" json:"activation_time,omitempty"`                            // 激活时间
	CreateTime      time.Time      `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP" json:"create_time"`           // 创建时间
	UpdateTime      time.Time      `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP" json:"update_time"`           // 更新时间
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`                                // 软删除标记
	Remark          string         `gorm:"column:remark;type:varchar(255)" json:"remark,omitempty"`                            // 备注

	// NetworkInfo DeviceNetwork `gorm:"foreignKey:DeviceID" json:"network_info,omitempty"` // 网络信息
	// StatusInfo  DeviceStatus  `gorm:"foreignKey:DeviceID" json:"status_info,omitempty"`  // 状态信息
}

// DeviceNetwork 设备网络信息模型
type DeviceNetwork struct {
	ID                 int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID           int64          `gorm:"column:device_id;not null;index" json:"device_id"`                                                   // 关联设备ID
	NetworkType        string         `gorm:"column:network_type;type:network_type_enum;not null" json:"network_type"`                            // 网络类型
	MacAddress         string         `gorm:"column:mac_address;type:varchar(64);uniqueIndex" json:"mac_address,omitempty"`                       // MAC地址
	IPAddress          string         `gorm:"column:ip_address;type:varchar(64)" json:"ip_address,omitempty"`                                     // IP地址
	Port               int            `gorm:"column:port" json:"port,omitempty"`                                                                  // 端口
	SignalStrength     int            `gorm:"column:signal_strength" json:"signal_strength,omitempty"`                                            // 信号强度
	ConnectStatus      string         `gorm:"column:connect_status;type:connect_status_enum;not null;default:disconnected" json:"connect_status"` // 连接状态
	LastConnectTime    time.Time      `gorm:"column:last_connect_time" json:"last_connect_time,omitempty"`                                        // 最后连接时间
	LastDisconnectTime time.Time      `gorm:"column:last_disconnect_time" json:"last_disconnect_time,omitempty"`                                  // 最后断开时间
	CreateTime         time.Time      `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP" json:"create_time"`                           // 创建时间
	UpdateTime         time.Time      `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP" json:"update_time"`                           // 更新时间
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`                                                // 软删除标记
}

// DeviceStatus 设备状态信息模型
type DeviceStatus struct {
	ID                int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID          int64          `gorm:"column:device_id;not null;index" json:"device_id"`                                                                    // 关联设备ID
	BatteryLevel      int            `gorm:"column:battery_level;not null;default:100;check:battery_level >= 0 AND battery_level <= 100" json:"battery_level"`    // 电池电量
	PowerStatus       string         `gorm:"column:power_status;type:power_status_enum;not null;default:power_on" json:"power_status"`                            // 供电状态
	WorkingStatus     string         `gorm:"column:working_status;type:working_status_enum;not null;default:idle" json:"working_status"`                          // 工作状态
	LastHeartbeatTime time.Time      `gorm:"column:last_heartbeat_time" json:"last_heartbeat_time,omitempty"`                                                     // 最后心跳时间
	LastMessageTime   time.Time      `gorm:"column:last_message_time" json:"last_message_time,omitempty"`                                                         // 最后消息时间
	CPUUsage          float64        `gorm:"column:cpu_usage;type:numeric(5,2);check:cpu_usage >= 0 AND cpu_usage <= 100" json:"cpu_usage,omitempty"`             // CPU使用率
	MemoryUsage       float64        `gorm:"column:memory_usage;type:numeric(5,2);check:memory_usage >= 0 AND memory_usage <= 100" json:"memory_usage,omitempty"` // 内存使用率
	ErrorCode         string         `gorm:"column:error_code;type:varchar(32)" json:"error_code,omitempty"`                                                      // 错误码
	CreateTime        time.Time      `gorm:"column:create_time;not null;default:CURRENT_TIMESTAMP" json:"create_time"`                                            // 创建时间
	UpdateTime        time.Time      `gorm:"column:update_time;not null;default:CURRENT_TIMESTAMP" json:"update_time"`                                            // 更新时间
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`                                                                 // 软删除标记
}

// TableName 设置 BaseDevice 的表名为 `device_base`
func (BaseDevice) TableName() string {
	return "device_base"
}

// TableName 设置 DeviceNetwork 的表名为 `device_network`
func (DeviceNetwork) TableName() string {
	return "device_network"
}

// TableName 设置 DeviceStatus 的表名为 `device_status`
func (DeviceStatus) TableName() string {
	return "device_status"
}
