package dto

type StartConversationRequest struct {
	AgentID    uint  `json:"agent_id" binding:"required"`
	PropertyID *uint `json:"property_id" binding:"omitempty"`
}

type SendMessageRequest struct {
	Body string `json:"body" binding:"required,min=1,max=2000"`
}
