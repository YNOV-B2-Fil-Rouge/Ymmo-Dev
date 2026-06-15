// Package models holds the GORM entities mapping the database tables.
package models

type Role struct {
	ID         uint8  `gorm:"primaryKey;column:id" json:"id"`
	Code       string `gorm:"column:code" json:"code"`
	Label      string `gorm:"column:label" json:"label"`
	IsInternal bool   `gorm:"column:is_internal" json:"is_internal"`
}

func (Role) TableName() string { return "roles" }
