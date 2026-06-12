package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// CreateConversation inserts a new conversation.
func (r *ConversationRepository) CreateConversation(c *models.Conversation) error {
	return r.db.Create(c).Error
}

// FindConversation returns a conversation by id, or nil.
func (r *ConversationRepository) FindConversation(id uint) (*models.Conversation, error) {
	var c models.Conversation
	err := r.db.First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindExisting looks for an existing thread between the same client, agent and
// property, so we reuse it instead of creating duplicates.
func (r *ConversationRepository) FindExisting(clientID, agentID uint, propertyID *uint) (*models.Conversation, error) {
	q := r.db.Where("client_id = ? AND agent_id = ?", clientID, agentID)
	if propertyID != nil {
		q = q.Where("property_id = ?", *propertyID)
	} else {
		q = q.Where("property_id IS NULL")
	}

	var c models.Conversation
	err := q.First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListForUser returns every conversation the user takes part in (as client or
// agent), newest first.
func (r *ConversationRepository) ListForUser(userID uint) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := r.db.
		Preload("Client").
		Preload("Agent").
		Preload("Property").
		Where("client_id = ? OR agent_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&conversations).Error
	return conversations, err
}

// AddMessage inserts a message into a conversation.
func (r *ConversationRepository) AddMessage(m *models.Message) error {
	return r.db.Create(m).Error
}

// ListMessages returns a conversation's messages in chronological order.
func (r *ConversationRepository) ListMessages(conversationID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// MarkRead flags as read every message in the conversation that the reader did
// not send.
func (r *ConversationRepository) MarkRead(conversationID, readerID uint) error {
	return r.db.Model(&models.Message{}).
		Where("conversation_id = ? AND sender_id <> ?", conversationID, readerID).
		Update("is_read", true).Error
}

// UnreadCount counts the user's unread messages across all their conversations.
func (r *ConversationRepository) UnreadCount(userID uint) (int64, error) {
	var n int64
	err := r.db.Model(&models.Message{}).
		Joins("JOIN conversations c ON c.id = messages.conversation_id").
		Where("(c.client_id = ? OR c.agent_id = ?)", userID, userID).
		Where("messages.sender_id <> ?", userID).
		Where("messages.is_read = ?", false).
		Count(&n).Error
	return n, err
}
