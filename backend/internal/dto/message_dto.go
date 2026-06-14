package dto

// StartConversationRequest opens (or reuses) a thread with an agent,
// optionally about a property. The current user is the client side.
type StartConversationRequest struct {
	AgentID    uint  `json:"agent_id" binding:"required"`
	PropertyID *uint `json:"property_id" binding:"omitempty"`
}

// SendMessageRequest is the body of a new message.
type SendMessageRequest struct {
	Body string `json:"body" binding:"required,min=1,max=2000"`
}
