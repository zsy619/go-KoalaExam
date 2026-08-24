package entity

import (
	"time"

	"gorm.io/gorm"
)

// Class 班级
type Class struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"size:64;not null" json:"name"`
	Grade        string         `gorm:"size:32" json:"grade"` // 年级
	DepartmentID *int64         `gorm:"index" json:"department_id"`
	TeacherID    *int64         `gorm:"index" json:"teacher_id"` // 班主任
	StudentCnt   int            `gorm:"default:0" json:"student_cnt"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Class) TableName() string { return "ke_class" }

// Department 组织/院系
type Department struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:64;not null" json:"name"`
	ParentID  int64          `gorm:"default:0;index" json:"parent_id"`
	Sort      int            `gorm:"default:0" json:"sort"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (d *Department) TableName() string { return "ke_department" }
