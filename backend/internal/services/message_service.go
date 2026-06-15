package services

import (
	"errors"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrReceiverNotFound     = errors.New("receiver not found")
	ErrInvalidParticipants = errors.New("invalid conversation participants")
	ErrNotParticipant = errors.New("not a participant of this conversation")
)

type MessageService struct {
	conversations *repositories.ConversationRepository
	users         *repositories.UserRepository
}

func NewMessageService(conversations *repositories.ConversationRepository, users *repositories.UserRepository) *MessageService {
	return &MessageService{conversations: conversations, users: users}
}

func (s *MessageService) StartConversation(initiatorID uint, initiatorRole string, req dto.StartConversationRequest) (*models.Conversation, error) {
	receiverID := req.AgentID
	if receiverID == initiatorID {
		return nil, ErrInvalidParticipants
	}

	receiver, err := s.users.FindByID(receiverID)
	if err != nil {
		return nil, err
	}
	if receiver == nil {
		return nil, ErrReceiverNotFound
	}
	receiverRole := ""
	if receiver.Role != nil {
		receiverRole = receiver.Role.Code
	}

	var clientID, agentID uint
	switch initiatorRole {
	case "AGENT":
		if receiverRole != "BUYER" {
			return nil, ErrInvalidParticipants
		}
		agentID, clientID = initiatorID, receiverID
	case "BUYER", "SELLER":
		if receiverRole != "AGENT" {
			return nil, ErrInvalidParticipants
		}
		clientID, agentID = initiatorID, receiverID
	default:
		return nil, ErrInvalidParticipants
	}

	existing, err := s.conversations.FindExisting(clientID, agentID, req.PropertyID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	conv := &models.Conversation{
		ClientID:   clientID,
		AgentID:    agentID,
		PropertyID: req.PropertyID,
	}
	if err := s.conversations.CreateConversation(conv); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *MessageService) List(userID uint) ([]models.Conversation, error) {
	return s.conversations.ListForUser(userID)
}

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

func (s *MessageService) GetMessages(userID, conversationID uint) ([]models.Message, error) {
	if err := s.ensureParticipant(userID, conversationID); err != nil {
		return nil, err
	}
	if err := s.conversations.MarkRead(conversationID, userID); err != nil {
		return nil, err
	}
	return s.conversations.ListMessages(conversationID)
}

func (s *MessageService) DeleteConversation(userID, conversationID uint) error {
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
	return s.conversations.SoftDelete(conv, userID)
}

func (s *MessageService) UnreadCount(userID uint) (int64, error) {
	return s.conversations.UnreadCount(userID)
}

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
