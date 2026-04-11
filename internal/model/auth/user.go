// Package auth 认证相关数据模型
package auth

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"column:username;type:varchar(64);not null;uniqueIndex" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(128);not null" json:"-"`
	Nickname     string         `gorm:"column:nickname;type:varchar(128)" json:"nickname"`
	Email        string         `gorm:"column:email;type:varchar(128);uniqueIndex" json:"email"`
	Phone        string         `gorm:"column:phone;type:varchar(32);uniqueIndex" json:"phone"`
	AvatarURL    string         `gorm:"column:avatar_url;type:varchar(256)" json:"avatar"`
	Status       int8           `gorm:"column:status;not null;default:1" json:"status"` // 1-启用, 0-禁用, 2-锁定
	LastLoginAt  *time.Time     `gorm:"column:last_login_at" json:"last_login_at"`
	LastLoginIP  string         `gorm:"column:last_login_ip;type:varchar(64)" json:"last_login_ip"`
	FailedAttempts int           `gorm:"column:failed_attempts;not null;default:0" json:"-"`
	LockedAt     *time.Time     `gorm:"column:locked_at" json:"-"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`

	// 关联字段
	Roles []Role `gorm:"many2many:auth.user_roles;" json:"roles,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "auth.users"
}

// BeforeUpdate Hook
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// IsActive 检查用户是否启用
func (u *User) IsActive() bool {
	return u.Status == 1
}

// IsLocked 检查用户是否被锁定
func (u *User) IsLocked() bool {
	return u.Status == 2 || (u.LockedAt != nil && time.Since(*u.LockedAt) < 15*time.Minute)
}
