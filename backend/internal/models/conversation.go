package models

import "time"

// Conversation maps the `conversations` table: a thread between a client and
// an agent, optionally about a property.
type Conversation struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	PropertyID *uint     `gorm:"column:property_id" json:"property_id,omitempty"`
	ClientID   uint      `gorm:"column:client_id" json:"client_id"`
	AgentID    uint      `gorm:"column:agent_id" json:"agent_id"`
	// Soft-delete flags (one per participant).
	ClientDeleted bool   `gorm:"column:client_deleted" json:"client_deleted"`
	AgentDeleted  bool   `gorm:"column:agent_deleted" json:"agent_deleted"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (Conversation) TableName() string { return "conversations" }
