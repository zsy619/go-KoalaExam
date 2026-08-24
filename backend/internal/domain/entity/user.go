package entity

import (
	"time"

	"gorm.io/gorm"
)

// User 用户实体（管理员/教师/学生）
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password     string         `gorm:"size:128;not null" json:"-"` // bcrypt 加密
	Nickname     string         `gorm:"size:64" json:"nickname"`
	Email        string         `gorm:"size:128;index" json:"email"`
	Phone        string         `gorm:"size:32;index" json:"phone"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	Role         int8           `gorm:"default:3;index" json:"role"`   // 1:超管 2:教师 3:学生
	Gender       int8           `gorm:"default:0" json:"gender"`       // 0:未知 1:男 2:女
	Status       int8           `gorm:"default:1;index" json:"status"` // 0:禁用 1:正常
	ClassID      *int64         `gorm:"index" json:"class_id"`         // 所属班级（学生可选）
	DepartmentID *int64         `gorm:"index" json:"department_id"`
	LastLoginAt  *time.Time     `gorm:"default:null" json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:64" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (u *User) TableName() string { return "ke_user" }
