package models

import "time"

// Favorite maps the `favorites` table: a many-to-many link between a user
// and a property (composite primary key user_id + property_id).
type Favorite struct {
	UserID     uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	PropertyID uint      `gorm:"primaryKey;column:property_id" json:"property_id"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Favorite) TableName() string { return "favorites" }
