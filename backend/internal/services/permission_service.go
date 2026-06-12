package services

import "ymmo/internal/repositories"

type PermissionService struct {
	perms *repositories.PermissionRepository
}

func NewPermissionService(perms *repositories.PermissionRepository) *PermissionService {
	return &PermissionService{perms: perms}
}

// Matrix returns the department access matrix.
func (s *PermissionService) Matrix() ([]repositories.PermissionRow, error) {
	return s.perms.ListMatrix()
}
