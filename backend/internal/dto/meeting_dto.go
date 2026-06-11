package dto

import "time"

// CreateMeetingRequest creates an internal meeting.
type CreateMeetingRequest struct {
	AgencyID       uint16    `json:"agency_id" binding:"required"`
	Title          string    `json:"title" binding:"required,min=3,max=150"`
	Description    string    `json:"description" binding:"omitempty"`
	StartAt        time.Time `json:"start_at" binding:"required"` // RFC3339
	EndAt          time.Time `json:"end_at" binding:"required"`   // RFC3339
	Location       string    `json:"location" binding:"omitempty,max=150"`
	ParticipantIDs []uint    `json:"participant_ids" binding:"omitempty"`
}
