package services

import (
	"errors"
	"time"

	"ymmo/internal/dto"
	"ymmo/internal/models"
	"ymmo/internal/repositories"
)

var (
	ErrMeetingNotFound    = errors.New("meeting not found")
	ErrNotOrganizer       = errors.New("only the organizer can do this")
	ErrInvalidMeetingTime = errors.New("end must be after start, and start must be in the future")
)

type MeetingService struct {
	meetings *repositories.MeetingRepository
}

func NewMeetingService(meetings *repositories.MeetingRepository) *MeetingService {
	return &MeetingService{meetings: meetings}
}

// Create schedules a meeting and links its participants.
func (s *MeetingService) Create(organizerID uint, req dto.CreateMeetingRequest) (*models.Meeting, error) {
	if !req.EndAt.After(req.StartAt) || req.StartAt.Before(time.Now()) {
		return nil, ErrInvalidMeetingTime
	}

	meeting := &models.Meeting{
		AgencyID:    req.AgencyID,
		OrganizerID: organizerID,
		Title:       req.Title,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
	}
	if req.Description != "" {
		meeting.Description = &req.Description
	}
	if req.Location != "" {
		meeting.Location = &req.Location
	}

	if err := s.meetings.Create(meeting); err != nil {
		return nil, err
	}
	if err := s.meetings.AddParticipants(meeting.ID, req.ParticipantIDs); err != nil {
		return nil, err
	}
	return s.meetings.FindByID(meeting.ID)
}

// List returns the user's meetings (organized or attended).
func (s *MeetingService) List(userID uint) ([]models.Meeting, error) {
	return s.meetings.ListForUser(userID)
}

// Delete removes a meeting, restricted to its organizer.
func (s *MeetingService) Delete(userID, meetingID uint) error {
	meeting, err := s.meetings.FindByID(meetingID)
	if err != nil {
		return err
	}
	if meeting == nil {
		return ErrMeetingNotFound
	}
	if meeting.OrganizerID != userID {
		return ErrNotOrganizer
	}
	return s.meetings.Delete(meetingID)
}
