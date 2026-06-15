package models

import "time"

const (
	PropertyStatusDraft         = "DRAFT"
	PropertyStatusPendingReview = "PENDING_REVIEW"
	PropertyStatusAvailable     = "AVAILABLE"
	PropertyStatusUnderOffer    = "UNDER_OFFER"
	PropertyStatusSold          = "SOLD"
	PropertyStatusWithdrawn     = "WITHDRAWN"
)

// IsPubliclyVisible reports whether a property is browsable by clients
// (catalogue / detail page). Draft, pending-review, sold and withdrawn
// listings are not, so they can't be favorited, contacted, etc.
func IsPubliclyVisible(status string) bool {
	return status == PropertyStatusAvailable || status == PropertyStatusUnderOffer
}

type Property struct {
	ID          uint     `gorm:"primaryKey;column:id" json:"id"`
	Reference   string   `gorm:"column:reference" json:"reference"`
	Title       string   `gorm:"column:title" json:"title"`
	Description *string  `gorm:"column:description" json:"description,omitempty"`
	CategoryID  uint8    `gorm:"column:category_id" json:"category_id"`
	Status      string   `gorm:"column:status" json:"status"`
	Price       float64  `gorm:"column:price" json:"price"`
	Area        float64  `gorm:"column:area" json:"area"`
	Rooms       *uint8   `gorm:"column:rooms" json:"rooms,omitempty"`
	Bedrooms    *uint8   `gorm:"column:bedrooms" json:"bedrooms,omitempty"`
	Bathrooms   *uint8   `gorm:"column:bathrooms" json:"bathrooms,omitempty"`
	Floor       *int8    `gorm:"column:floor" json:"floor,omitempty"`
	BuildYear   *uint16  `gorm:"column:build_year" json:"build_year,omitempty"`
	EnergyRating *string `gorm:"column:energy_rating" json:"energy_rating,omitempty"`
	GhgRating   *string  `gorm:"column:ghg_rating" json:"ghg_rating,omitempty"`
	Address     *string  `gorm:"column:address" json:"address,omitempty"`
	City        string   `gorm:"column:city" json:"city"`
	PostalCode  string   `gorm:"column:postal_code" json:"postal_code"`
	Latitude    *float64 `gorm:"column:latitude" json:"latitude,omitempty"`
	Longitude   *float64 `gorm:"column:longitude" json:"longitude,omitempty"`
	IsExclusive bool     `gorm:"column:is_exclusive" json:"is_exclusive"`
	ViewCount   uint     `gorm:"column:view_count" json:"view_count"`
	AgencyID    uint16   `gorm:"column:agency_id" json:"agency_id"`
	AgentID     *uint    `gorm:"column:agent_id" json:"agent_id,omitempty"`
	SellerID    *uint    `gorm:"column:seller_id" json:"seller_id,omitempty"`
	ApprovedBy  *uint    `gorm:"column:approved_by" json:"approved_by,omitempty"`
	ApprovedAt  *time.Time `gorm:"column:approved_at" json:"approved_at,omitempty"`
	PublishedAt *time.Time `gorm:"column:published_at" json:"published_at,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	Category *PropertyCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Photos   []PropertyPhoto   `gorm:"foreignKey:PropertyID" json:"photos,omitempty"`
}

func (Property) TableName() string { return "properties" }
