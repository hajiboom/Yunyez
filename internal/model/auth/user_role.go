package auth

import "time"

// UserRole 用户角色关联模型
type UserRole struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:idx_user_role" json:"user_id"`
	RoleID    int64     `gorm:"column:role_id;not null;uniqueIndex:idx_user_role" json:"role_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	CreatedBy int64     `gorm:"column:created_by" json:"created_by"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "auth.user_roles"
}
