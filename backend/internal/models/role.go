// Package models contains the GORM entities. Each struct maps an
// EXISTING table from db/schema.sql (the SQL schema is the source of
// truth). We deliberately do NOT use AutoMigrate so the hand-written
// constraints, indexes and ENUMs stay authoritative.
package models

// Role is an application role (RBAC). Maps the `roles` table.
type Role struct {
	ID         uint8  `gorm:"primaryKey;column:id" json:"id"`
	Code       string `gorm:"column:code" json:"code"`   // VISITOR, BUYER, AGENT, ...
	Label      string `gorm:"column:label" json:"label"`
	IsInternal bool   `gorm:"column:is_internal" json:"is_internal"`
}

// TableName pins the table name (GORM would otherwise pluralize to "roles"
// anyway, but being explicit avoids surprises).
func (Role) TableName() string { return "roles" }
