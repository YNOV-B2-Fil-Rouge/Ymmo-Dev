package models

import "time"

// User maps the `users` table. A single table holds every actor
// (visitor/buyer/seller/agent/...), distinguished by RoleID.
// DepartmentID and AgencyID only apply to internal staff -> pointers
// (nullable).
type User struct {
	ID           uint    `gorm:"primaryKey;column:id" json:"id"`
	Email        string  `gorm:"column:email" json:"email"`
	PasswordHash string  `gorm:"column:password_hash" json:"-"` // never serialized to JSON
	LastName     string  `gorm:"column:last_name" json:"last_name"`
	FirstName    string  `gorm:"column:first_name" json:"first_name"`
	Phone        *string `gorm:"column:phone" json:"phone,omitempty"`
	RoleID       uint8   `gorm:"column:role_id" json:"role_id"`
	DepartmentID *uint8  `gorm:"column:department_id" json:"department_id,omitempty"`
	AgencyID     *uint16 `gorm:"column:agency_id" json:"agency_id,omitempty"`
	IsActive     bool    `gorm:"column:is_active" json:"is_active"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Association loaded on demand with Preload("Role").
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

func (User) TableName() string { return "users" }
