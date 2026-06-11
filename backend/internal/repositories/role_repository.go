package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// FindByCode returns the role matching a code (e.g. "BUYER") or nil.
func (r *RoleRepository) FindByCode(code string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}
