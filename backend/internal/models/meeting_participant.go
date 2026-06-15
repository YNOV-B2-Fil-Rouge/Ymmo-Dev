package models

type MeetingParticipant struct {
	MeetingID uint `gorm:"primaryKey;column:meeting_id" json:"meeting_id"`
	UserID    uint `gorm:"primaryKey;column:user_id" json:"user_id"`
}

func (MeetingParticipant) TableName() string { return "meeting_participants" }
