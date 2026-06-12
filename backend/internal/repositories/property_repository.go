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

// Create inserts a property and assigns its human-readable reference
// (e.g. YMMO-2026-00042) based on the generated ID. Done in a transaction
// so a failure never leaves a half-written row.
func (r *PropertyRepository) Create(p *models.Property) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Temporary unique reference to satisfy the NOT NULL/UNIQUE column
		// before we know the auto-increment ID. Kept short to fit
		// reference VARCHAR(20): "TMP-" + 11 digits = 15 chars.
		p.Reference = fmt.Sprintf("TMP-%011d", time.Now().UnixNano()%100000000000)
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		p.Reference = fmt.Sprintf("YMMO-%d-%05d", time.Now().Year(), p.ID)
		return tx.Model(p).Update("reference", p.Reference).Error
	})
}

// FindByID returns a property with its category and photos, or nil.
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

// Update applies a partial set of columns to a property.
func (r *PropertyRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Property{}).Where("id = ?", id).Updates(updates).Error
}

// Delete removes a property (its photos cascade via the FK).
func (r *PropertyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Property{}, id).Error
}

// ListByAgent returns every property assigned to an agent, ALL statuses
// (drafts, sold, ...) — used by the agent dashboard "My properties".
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

// ListAll returns every property, ALL statuses — for the director/HQ
// management view.
func (r *PropertyRepository) ListAll() ([]models.Property, error) {
	var properties []models.Property
	err := r.db.
		Preload("Category").
		Preload("Photos").
		Order("created_at DESC").
		Find(&properties).Error
	return properties, err
}

// IncrementViewCount bumps the fast popularity counter by one.
func (r *PropertyRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&models.Property{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// Search runs the public catalogue query: dynamic filters + pagination.
// Returns the page of results and the total count (for pagination metadata).
func (r *PropertyRepository) Search(q dto.PropertySearchQuery) ([]models.Property, int64, error) {
	query := r.db.Model(&models.Property{})

	// Filter by category sector requires the categories table.
	if q.Sector != "" {
		query = query.Joins("JOIN property_categories pc ON pc.id = properties.category_id").
			Where("pc.sector = ?", q.Sector)
	}

	if q.City != "" {
		query = query.Where("properties.city = ?", q.City)
	}
	if q.CategoryID != 0 {
		query = query.Where("properties.category_id = ?", q.CategoryID)
	}
	if q.MinPrice > 0 {
		query = query.Where("properties.price >= ?", q.MinPrice)
	}
	if q.MaxPrice > 0 {
		query = query.Where("properties.price <= ?", q.MaxPrice)
	}
	if q.MinArea > 0 {
		query = query.Where("properties.area >= ?", q.MinArea)
	}
	if q.MaxArea > 0 {
		query = query.Where("properties.area <= ?", q.MaxArea)
	}
	if q.MaxEnergy != "" {
		// Letters A..G are ordered, so "<=" returns this rating or better.
		query = query.Where("properties.energy_rating <= ?", q.MaxEnergy)
	}

	// Status: default to AVAILABLE so the public never sees drafts/sold items.
	status := q.Status
	if status == "" {
		status = models.PropertyStatusAvailable
	}
	query = query.Where("properties.status = ?", status)

	// Count before applying limit/offset.
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting.
	switch q.Sort {
	case "price_asc":
		query = query.Order("properties.price ASC")
	case "price_desc":
		query = query.Order("properties.price DESC")
	default: // "recent"
		query = query.Order("properties.created_at DESC")
	}

	// Pagination with sane bounds.
	page := q.Page
	if page < 1 {
		page = 1
	}
	size := q.PageSize
	if size < 1 {
		size = 12
	}
	if size > 50 {
		size = 50
	}

	var properties []models.Property
	err := query.
		Preload("Category").
		Preload("Photos").
		Limit(size).
		Offset((page - 1) * size).
		Find(&properties).Error
	if err != nil {
		return nil, 0, err
	}
	return properties, total, nil
}
