package auth

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleCode    string         `gorm:"column:role_code;type:varchar(64);not null;uniqueIndex" json:"role_code"`
	RoleName    string         `gorm:"column:role_name;type:varchar(128);not null" json:"role_name"`
	Description string         `gorm:"column:description;type:varchar(256)" json:"description"`
	Status      int8           `gorm:"column:status;not null;default:1" json:"status"` // 1-启用, 0-禁用
	CreatedAt   time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "auth.roles"
}

// BeforeUpdate Hook
func (r *Role) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// IsActive 检查角色是否启用
func (r *Role) IsActive() bool {
	return r.Status == 1
}
