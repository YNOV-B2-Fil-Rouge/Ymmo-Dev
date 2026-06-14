package models

import "time"

// Meeting maps the `meetings` table: an internal team meeting organized by a
// staff member, optionally with participants.
type Meeting struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	AgencyID    uint16    `gorm:"column:agency_id" json:"agency_id"`
	OrganizerID uint      `gorm:"column:organizer_id" json:"organizer_id"`
	Title       string    `gorm:"column:title" json:"title"`
	Description *string   `gorm:"column:description" json:"description,omitempty"`
	StartAt     time.Time `gorm:"column:start_at" json:"start_at"`
	EndAt       time.Time `gorm:"column:end_at" json:"end_at"`
	Location    *string   `gorm:"column:location" json:"location,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`

	// Read side of the N-N link (write goes through MeetingParticipant).
	Participants []User `gorm:"many2many:meeting_participants;" json:"participants,omitempty"`
}

func (Meeting) TableName() string { return "meetings" }
