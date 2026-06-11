package models

// PropertyCategory maps the `property_categories` table
// (sector RESIDENTIAL/COMMERCIAL + a label like House, Office...).
type PropertyCategory struct {
	ID     uint8  `gorm:"primaryKey;column:id" json:"id"`
	Sector string `gorm:"column:sector" json:"sector"`
	Label  string `gorm:"column:label" json:"label"`
}

func (PropertyCategory) TableName() string { return "property_categories" }
