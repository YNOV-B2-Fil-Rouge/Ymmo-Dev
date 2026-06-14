// Package repositories isolates all data access. Services depend on these
// types, never on *gorm.DB directly -> the persistence layer can change
// without touching business logic (Dependency Inversion).
package repositories

import (
	"errors"

	"gorm.io/gorm"

	"ymmo/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user row.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail returns the user (with its Role preloaded) or nil if absent.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID returns the user (with its Role) or nil if absent.
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateRole changes a user's role (used to promote a buyer to seller after an
// agent approves their application).
func (r *UserRepository) UpdateRole(userID uint, roleID uint8) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("role_id", roleID).Error
}

// ListInternal returns the company's internal staff (the collaborator
// directory), with their role preloaded.
func (r *UserRepository) ListInternal() ([]models.User, error) {
	var users []models.User
	err := r.db.
		Preload("Role").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.is_internal = ?", true).
		Order("users.last_name").
		Find(&users).Error
	return users, err
}

// ListAll returns every user (for IT user management), role preloaded.
func (r *UserRepository) ListAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Role").Order("users.last_name").Find(&users).Error
	return users, err
}

// ListInternalByAgency returns the internal staff of a single agency.
func (r *UserRepository) ListInternalByAgency(agencyID uint16) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Preload("Role").
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.is_internal = ? AND users.agency_id = ?", true, agencyID).
		Order("users.last_name").
		Find(&users).Error
	return users, err
}

// ExistsByEmail reports whether an account already uses this email.
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
