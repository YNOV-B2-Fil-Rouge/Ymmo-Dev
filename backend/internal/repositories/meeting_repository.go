package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ymmo/internal/models"
)

type MeetingRepository struct {
	db *gorm.DB
}

func NewMeetingRepository(db *gorm.DB) *MeetingRepository {
	return &MeetingRepository{db: db}
}

func (r *MeetingRepository) Create(m *models.Meeting) error {
	return r.db.Omit("Participants").Create(m).Error
}

func (r *MeetingRepository) AddParticipants(meetingID uint, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	rows := make([]models.MeetingParticipant, 0, len(userIDs))
	for _, uid := range userIDs {
		rows = append(rows, models.MeetingParticipant{MeetingID: meetingID, UserID: uid})
	}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *MeetingRepository) ExistsAtTime(organizerID uint, startAt time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.Meeting{}).
		Where("organizer_id = ? AND start_at = ?", organizerID, startAt).
		Count(&count).Error
	return count > 0, err
}

func (r *MeetingRepository) FindByID(id uint) (*models.Meeting, error) {
	var m models.Meeting
	err := r.db.Preload("Participants").First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MeetingRepository) ListForUser(userID uint) ([]models.Meeting, error) {
	attended := r.db.Table("meeting_participants").
		Select("meeting_id").Where("user_id = ?", userID)

	var meetings []models.Meeting
	err := r.db.
		Where("organizer_id = ?", userID).
		Or("id IN (?)", attended).
		Preload("Participants").
		Order("start_at ASC").
		Find(&meetings).Error
	return meetings, err
}

func (r *MeetingRepository) Delete(id uint) error {
	return r.db.Delete(&models.Meeting{}, id).Error
}
