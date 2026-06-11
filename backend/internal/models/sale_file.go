package models

import "time"

// Sale file status values (must match the ENUM in db/schema.sql).
const (
	SaleStatusOffer               = "OFFER"
	SaleStatusPreliminaryContract = "PRELIMINARY_CONTRACT" // "compromis"
	SaleStatusDeed                = "DEED"                  // "acte authentique"
	SaleStatusCompleted           = "COMPLETED"
	SaleStatusCancelled           = "CANCELLED"
)

// SaleFile maps the `sale_files` table: the lifecycle of a transaction
// between a buyer and an agent for a property.
type SaleFile struct {
	ID              uint       `gorm:"primaryKey;column:id" json:"id"`
	PropertyID      uint       `gorm:"column:property_id" json:"property_id"`
	BuyerID         uint       `gorm:"column:buyer_id" json:"buyer_id"`
	AgentID         uint       `gorm:"column:agent_id" json:"agent_id"`
	Status          string     `gorm:"column:status" json:"status"`
	NegotiatedPrice *float64   `gorm:"column:negotiated_price" json:"negotiated_price,omitempty"`
	OfferDate       *time.Time `gorm:"column:offer_date" json:"offer_date,omitempty"`
	ContractDate    *time.Time `gorm:"column:contract_date" json:"contract_date,omitempty"`
	DeedDate        *time.Time `gorm:"column:deed_date" json:"deed_date,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (SaleFile) TableName() string { return "sale_files" }
