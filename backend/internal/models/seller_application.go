package models

import "time"

// Seller application status values (must match the ENUM in db/schema.sql).
const (
	SellerApplicationPending  = "PENDING"
	SellerApplicationApproved = "APPROVED"
	SellerApplicationRejected = "REJECTED"
)

// SellerApplication maps the `seller_applications` table: a buyer's request to
// be promoted to seller, reviewed by an agent. Approval changes the user's role.
type SellerApplication struct {
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint       `gorm:"column:user_id" json:"user_id"`
	Status     string     `gorm:"column:status" json:"status"`
	Motivation *string    `gorm:"column:motivation" json:"motivation,omitempty"`
	ReviewedBy *uint      `gorm:"column:reviewed_by" json:"reviewed_by,omitempty"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	ReviewedAt *time.Time `gorm:"column:reviewed_at" json:"reviewed_at,omitempty"`

	// Eager-loaded applicant, for the agent's review list.
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SellerApplication) TableName() string { return "seller_applications" }
