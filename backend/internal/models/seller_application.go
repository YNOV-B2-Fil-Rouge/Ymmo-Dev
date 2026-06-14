package models

import "time"

const (
	SellerApplicationPending  = "PENDING"
	SellerApplicationApproved = "APPROVED"
	SellerApplicationRejected = "REJECTED"
)

type SellerApplication struct {
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint       `gorm:"column:user_id" json:"user_id"`
	Status     string     `gorm:"column:status" json:"status"`
	Motivation *string    `gorm:"column:motivation" json:"motivation,omitempty"`
	ReviewedBy *uint      `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	ReviewedAt *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SellerApplication) TableName() string { return "seller_applications" }
