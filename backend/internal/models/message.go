package models

import "time"

type Message struct {
	ID             uint      `gorm:"primaryKey;column:id" json:"id"`
	ConversationID uint      `gorm:"column:conversation_id" json:"conversation_id"`
	SenderID       uint      `gorm:"column:sender_id" json:"sender_id"`
	Body           string    `gorm:"column:body" json:"body"`
	IsRead         bool      `gorm:"column:is_read" json:"is_read"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Message) TableName() string { return "messages" }
