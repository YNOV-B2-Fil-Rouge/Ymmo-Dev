package models

// Department maps the `departments` table (the company's internal poles
// used by the access matrix).
type Department struct {
	ID          uint8   `gorm:"primaryKey;column:id" json:"id"`
	Name        string  `gorm:"column:name" json:"name"`
	Description *string `gorm:"column:description" json:"description,omitempty"`
}

func (Department) TableName() string { return "departments" }
