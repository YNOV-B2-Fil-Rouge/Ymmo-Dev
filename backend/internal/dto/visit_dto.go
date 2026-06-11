package dto

import "time"

// RequestVisitRequest is a client's request to visit a property.
type RequestVisitRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"` // RFC3339, e.g. 2026-07-01T15:00:00Z
	Notes       string    `json:"notes" binding:"omitempty,max=500"`
}

// UpdateVisitStatusRequest changes a visit's status.
type UpdateVisitStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=REQUESTED CONFIRMED CANCELLED COMPLETED"`
}
