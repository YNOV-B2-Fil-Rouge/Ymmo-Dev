package repositories

import "gorm.io/gorm"

type PermissionRow struct {
	Requester string `json:"requester"`
	Target    string `json:"target"`
	Level     string `json:"level"`
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) ListMatrix() ([]PermissionRow, error) {
	var rows []PermissionRow
	err := r.db.
		Table("department_permissions AS p").
		Select("dr.name AS requester, dt.name AS target, p.access_level AS level").
		Joins("JOIN departments dr ON dr.id = p.requester_department_id").
		Joins("JOIN departments dt ON dt.id = p.target_department_id").
		Order("p.requester_department_id, p.target_department_id").
		Scan(&rows).Error
	return rows, err
}
