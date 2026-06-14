package dto

import "time"

type RequestVisitRequest struct {
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
	Notes       string    `json:"notes" binding:"omitempty,max=500"`
}

type UpdateVisitStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=REQUESTED CONFIRMED CANCELLED COMPLETED"`
}
