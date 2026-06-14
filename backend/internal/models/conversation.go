package models

import "time"

type Conversation struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	PropertyID *uint     `gorm:"column:property_id" json:"property_id,omitempty"`
	ClientID   uint      `gorm:"column:client_id" json:"client_id"`
	AgentID    uint      `gorm:"column:agent_id" json:"agent_id"`
	ClientDeleted bool   `gorm:"column:client_deleted" json:"client_deleted"`
	AgentDeleted  bool   `gorm:"column:agent_deleted" json:"agent_deleted"`
	CreatedAt  time.Time `gorm: