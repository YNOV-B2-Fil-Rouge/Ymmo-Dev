package repositories

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ymmo/internal/dto"
	"ymmo/internal/models"
)

type PropertyRepository struct {
	db *gorm.DB
}

func NewPropertyRepository(db *gorm.DB) *PropertyRepository {
	return &PropertyRepository{db: db}
}

func (r *PropertyRepository) Create(p *models.Property) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		p.Reference = fmt.Sprintf("TMP-%011d", time.Now().UnixNano()%100000000000)
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		p.Reference = fmt.Sprintf("YMMO-%d-%05d", time.Now().Year(), p.ID)
		return tx.Model(p).Update("reference", p.Reference).Error
	})
}

func (r *PropertyRepository) FindByID(id uint) (*models.Property, error) {
	var p models.Property
	err := r.db.Preload("Category").Preload("Photos").First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PropertyRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Property{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PropertyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Property{}, id).Error
}

func (r *PropertyRepository) ListByAgent(agentID uint) ([]models.Property, error) {
	var properties []models.Property
	err := r.db.
		Where("agent_id = ?", agentID).
		Preload("Category").
		Preload("Photos").
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

func (r *PropertyRepository) ListByAgency(agencyID uint16) ([]models.Property, error) {
	var properties []models.Property
	err := r.db.
		Where("agency_id = ?", agencyID).
		Preload("Category").
		Preload("Photos").
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

func (r *PropertyRepository) ListAll() ([]models.Property, error) {
	var properties []models.Property
	err := r.db.
		Preload("Category").
		Preload("Photos").
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

func (r *PropertyRepository) ListPending(agencyID *uint16) ([]models.Property, error) {
	var properties []models.Property
	q := r.db.
		Where("status = ?", models.PropertyStatusPendingReview).
		Preload("Category").
		Preload("Photos").
		Order("created_at DESC")
	if agencyID != nil {
		q = q.Where("agency_id = ?", *agencyID)
	}
	err := q.Find(&properties).Error
	return properties, err
}

func (r *PropertyRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&models.Property{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PropertyRepository) Search(q dto.PropertySearchQuery) ([]models.Property, int64, error) {
	query := r.db.Model(&models.Property{})

	if q.Sector != "" {
		query