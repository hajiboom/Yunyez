package auth

import "time"

// LoginLog 登录日志模型
type LoginLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        *int64    `gorm:"column:user_id" json:"user_id"`
	Username      string    `gorm:"column:username;type:varchar(64);not null" json:"username"`
	LoginType     string    `gorm:"column:login_type;type:varchar(32);not null;default:password" json:"login_type"`
	Status        int8      `gorm:"column:status;not null" json:"status"` // 1-成功, 0-失败
	FailureReason string    `gorm:"column:failure_reason;type:varchar(128)" json:"failure_reason"`
	IPAddress     string    `gorm:"column:ip_address;type:varchar(64)" json:"ip_address"`
	UserAgent     string    `gorm:"column:user_agent;type:varchar(256)" json:"user_agent"`
	DeviceInfo    string    `gorm:"column:device_info;type:varchar(256)" json:"device_info"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名
func (LoginLog) TableName() string {
	return "auth.login_logs"
}
