package models

import "time"

type Alert struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	UserID     uint      `gorm:"column:user_id" json:"user_id"`
	City       *string   `gorm:"column:city" json:"city,omitempty"`
	CategoryID *uint8    `gorm:"column:category_id" json:"category_id,omitempty"`
	MinPrice   *float64  `gorm:"column:min_price" json:"min_price,omitempty"`
	MaxPrice   *float64  `gorm:"column:max_price" json:"max_price,omitempty"`
	MinArea    *float64  `gorm:"column:min_area" json:"min_area,omitempty"`
	MaxEnergy  *string   `gorm:"column:max_energy" json:"max_energy,omitempty"`
	IsActive   bool      `gorm:"column:is_active" json:"is_active"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Alert) TableName() string { return "alerts" }
