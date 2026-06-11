package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrAgentNotFound        = errors.New("agent not found")
	// ErrNotParticipant guards access: a user may only touch conversations
	// they are part of (the brief's "secure messaging" requirement).
	ErrNotParticipant = errors.New("not a participant of this conversation")
)

type MessageService struct {
	conversations *repositories.ConversationRepository
	users         *repositories.UserRepository
}

func NewMessageService(conversations *repositories.ConversationRepository, users *repositories.UserRepository) *MessageService {
	return &MessageService{conversations: conversations, users: users}
}

// StartConversation opens a thread between the current user (client) and an
// agent. If an identical thread already exists, it is returned instead of
// creating a duplicate.
func (s *MessageService) StartConversation(clientID uint, req dto.StartConversationRequest) (*models.Conversation, error) {
	agent, err := s.users.FindByID(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	existing, err := s.conversations.FindExisting(clientID, req.AgentID, req.PropertyID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	conv := &models.Conversation{
		ClientID:   clientID,
		AgentID:    req.AgentID,
		PropertyID: req.PropertyID,
	}
	if err := s.conversations.CreateConversation(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

// List returns the user's conversations.
func (s *MessageService) List(userID uint) ([]models.Conversation, error) {
	return s.conversations.ListForUser(userID)
}

// SendMessage posts a message, after checking the user belongs to the thread.
func (s *MessageService) SendMessage(userID, conversationID uint, body string) (*models.Message, error) {
	if err := s.ensureParticipant(userID, conversationID); err != nil {
		return nil, err
	}
	msg := &models.Message{
		ConversationID: conversationID,
		SenderID:       userID,
		Body:           body,
	}
	if err := s.conversations.AddMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// GetMessages returns a thread's messages and marks the others as read.
func (s *MessageService) GetMessages(userID, conversationID uint) ([]models.Message, error) {
	if err := s.ensureParticipant(userID, conversationID); err != nil {
		return nil, err
	}
	if err := s.conversations.MarkRead(conversationID, userID); err != nil {
		return nil, err
	}
	return s.conversations.ListMessages(conversationID)
}

// UnreadCount returns how many unread messages the user has.
func (s *MessageService) UnreadCount(userID uint) (int64, error) {
	return s.conversations.UnreadCount(userID)
}

// ensureParticipant returns ErrConversationNotFound or ErrNotParticipant when
// the user is not allowed to access the conversation.
func (s *MessageService) ensureParticipant(userID, conversationID uint) error {
	conv, err := s.conversations.FindConversation(conversationID)
	if err != nil {
		return err
	}
	if conv == nil {
		return ErrConversationNotFound
	}
	if conv.ClientID != userID && conv.AgentID != userID {
		return ErrNotParticipant
	}
	return nil
}
