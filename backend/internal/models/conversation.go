package models

import "time"

type Conversation struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	PropertyID *uint     `gorm:"column:property_id" json:"property_id,omitempty"`
	ClientID   uint      `gorm:"column:client_id" json:"client_id"`
	AgentID    uint      `gorm:"column:agent_id" json:"agent_id"`
	ClientDeleted bool   `gorm:"column:client_deleted" json:"client_deleted"`
	AgentDeleted  bool   `gorm:"column:agent_deleted" json:"agent_deleted"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`

	Client   *User     `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Agent    *User     `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
	Property *Property `gorm:"foreignKey:PropertyID" json:"property,omitempty"`

	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

func (Conversation) TableName() string { return "conversations" }
