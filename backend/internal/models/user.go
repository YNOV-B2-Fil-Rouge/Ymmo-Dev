package models

import "time"

type User struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	Email        string  `gorm:"column:email" json:"email"`
	PasswordHash string  `gorm:"column:password_hash" json:"-"`
	LastName     string  `gorm:"column:last_name" json:"last_name"`
	FirstName    string  `gorm:"column:first_name" json:"first_name"`
	Phone        *string `gorm:"column:phone" json:"phone,omitempty"`
	RoleID       uint8   `gorm:"column:role_id" json:"role_id"`
	DepartmentID *uint8  `gorm:"column:department_id" json:"department_id,omitempty"`
	AgencyID     *uint16 `gorm:"column:agency_id" json:"agency_id,omitempty"`
	IsActive     bool    `gorm:"column:is_active" json:"is_active"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (User) TableName() string { return "users" }
