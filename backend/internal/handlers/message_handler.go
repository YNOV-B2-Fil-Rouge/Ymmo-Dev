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

	initiatorID := middleware.CurrentUserID(c)
	initiatorRole := middleware.CurrentRole(c)
	conv, err := h.svc.StartConversation(initiatorID, initiatorRole, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrReceiverNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "receiver not found"})
		case errors.Is(err, services.ErrInvalidParticipants):
			c.JSON(http.StatusBadRequest, gin.H{"error": "a conversation must be between a client and an agent"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start conversation"})
		}
		return
	}
	c.JSON(http.StatusCreated, conv)
}

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

// @Summary      Delete a conversation
// @Tags         messaging
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Conversation id"
// @Success      204  "No Content"
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /conversations/{id} [delete]
func (h *MessageHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID := middleware.CurrentUserID(c)
	if err := h.svc.DeleteConversation(userID, id); err != nil {
		h.writeAccessError(c, err, "could not delete conversation")
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      Count my unread messages
// @Tags         messaging
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]int
// @Failure      401  {object}  map[string]string
// @Router       /messages/unread-count [get]
func (h *MessageHandler) UnreadCount(c *gin.Context