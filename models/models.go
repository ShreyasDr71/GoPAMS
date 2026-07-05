package models

import (
	"time"

	"gorm.io/gorm"
)

// Group represents an organizational group (e.g., Development, Finance)
type Group struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
}

// Role represents a user role with hierarchical privileges (e.g., Administrator, Guest)
type Role struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Name           string         `gorm:"uniqueIndex;not null" json:"name"`
	Description    string         `json:"description"`
	HierarchyLevel int            `gorm:"not null;default:0" json:"hierarchy_level"` // Higher number = more privilege
}

// User represents a system user account
type User struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	FullName           string         `gorm:"not null" json:"full_name"`
	Username           string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash       string         `gorm:"not null" json:"-"`
	PhoneNumber        string         `json:"phone_number"`
	Email              *string        `json:"email"`
	EmployeeID         *string        `json:"employee_id"`
	Status             string         `gorm:"type:varchar(20);not null;default:'Active'" json:"status"`
	LastLoginAt        *time.Time     `json:"last_login_at"`
	GroupID            *uint          `json:"group_id"`
	Group              *Group         `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	RoleID             *uint          `json:"role_id"`
	Role               *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	MustChangePassword bool           `gorm:"default:true" json:"must_change_password"`
	IsAdmin            bool           `gorm:"default:false" json:"is_admin"`
}
