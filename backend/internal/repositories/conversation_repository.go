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

func (r *ConversationRepository) CreateConversation(c *models.Conversation) error {
	return r.db.Create(c).Error
}

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

func (r *ConversationRepository) ListForUser(userID uint) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := r.db.
		Preload("Client").
		Preload("Agent").
		Preload("Property").
		Where("(client_id = ? AND client_deleted = ?) OR (agent_id = ? AND agent_deleted = ?)",
			userID, false, userID, false).
		Order("created_at DESC").
		Find(&conversations).Error
	return conversations, err
}

func (r *ConversationRepository) SoftDelete(conv *models.Conversation, userID uint) error {
	if conv.ClientID == userID {
		conv.ClientDeleted = true
	} else if conv.AgentID == userID {
		conv.AgentDeleted = true
	}

	if conv.ClientDeleted && conv.AgentDeleted {
		return r.db.Delete(&models.Conversation{}, conv.ID).Error
	}
	return r.db.Model(conv).Updates(map[string]interface{}{
		"client_deleted": conv.ClientDeleted,
		"agent_deleted":  conv.AgentDeleted,
	}).Error
}

func (r *ConversationRepository) AddMessage(m *models.Message) error {
	return r.db.Create(m).Error
}

func (r *ConversationRepository) ListMessages(conversationID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}
