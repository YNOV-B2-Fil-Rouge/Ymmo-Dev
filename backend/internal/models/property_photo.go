package models

type PropertyPhoto struct {
	ID         uint   `gorm:"primaryKey;column:id" json:"id"`
	PropertyID uint   `gorm:"column:property_id" json:"property_id"`
	URL        string `gorm:"column:url" json:"url"`
	SortOrder  uint8  `gorm:"column:sort_order" json:"sort_order"`
	IsPrimary  bool   `gorm:"column:is_primary" json:"is_primary"`
}

func (PropertyPhoto) TableName() string { return "property_photos" }
