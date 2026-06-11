package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"ymmo/internal/dto"
	"ymmo/internal/middleware"
	"ymmo/internal/services"
)

type MessageHandler struct {
	svc *services.MessageService
}

func NewMessageHandler(svc *services.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

// StartConversation opens (or reuses) a thread with an agent.
// @Summary      Start a conversation with an agent
// @Tags         messaging
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.StartConversationRequest  true  "Agent (and optional property)"
// @Success      201   {object}  models.Conversation
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /conversations [post]
func (h *MessageHandler) StartConversation(c *gin.Context) {
	var req dto.StartConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	clientID := middleware.CurrentUserID(c)
	conv, err := h.svc.StartConversation(clientID, req)
	if err != nil {
		if errors.Is(err, services.ErrAgentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start conversation"})
		return
	}
	c.JSON(http.StatusCreated, conv)
}

// ListConversations returns the current user's threads.
// @Summary      List my conversations
// @Tags         messaging
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Router       /conversations [get]
func (h *MessageHandler) ListConversations(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	items, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch conversations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// SendMessage posts a message in a conversation.
// @Summary      Send a message
// @Tags         messaging
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                     true  "Conversation id"
// @Param        body  body      dto.SendMessageRequest  true  "Message body"
// @Success      201   {object}  models.Message
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /conversations/{id}/messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	userID := middleware.CurrentUserID(c)
	msg, err := h.svc.SendMessage(userID, id, req.Body)
	if err != nil {
		h.writeAccessError(c, err, "could not send message")
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// GetMessages returns a conversation's messages (and marks them read).
// @Summary      Get conversation messages
// @Tags         messaging
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Conversation id"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /conversations/{id}/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	userID := middleware.CurrentUserID(c)
	messages, err := h.svc.GetMessages(userID, id)
	if err != nil {
		h.writeAccessError(c, err, "could not fetch messages")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// UnreadCount returns the user's number of unread messages.
// @Summary      Count my unread messages
// @Tags         messaging
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]int
// @Failure      401  {object}  map[string]string
// @Router       /messages/unread-count [get]
func (h *MessageHandler) UnreadCount(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	count, err := h.svc.UnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count unread messages"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unread": count})
}

// writeAccessError maps the messaging access errors to HTTP status codes.
func (h *MessageHandler) writeAccessError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, services.ErrConversationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
	case errors.Is(err, services.ErrNotParticipant):
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a participant of this conversation"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}
