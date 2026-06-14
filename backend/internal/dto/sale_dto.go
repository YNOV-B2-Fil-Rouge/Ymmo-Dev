package dto

import "time"

// CreateSaleRequest opens a sale file for a property and a buyer.
type CreateSaleRequest struct {
	PropertyID      uint       `json:"property_id" binding:"required"`
	BuyerID         uint       `json:"buyer_id" binding:"required"`
	NegotiatedPrice *float64   `json:"negotiated_price" binding:"omitempty,gte=0"`
	OfferDate       *time.Time `json:"offer_date" binding:"omitempty"`
}

// UpdateSaleRequest advances a sale file. All fields are optional (partial).
type UpdateSaleRequest struct {
	Status          *string    `json:"status" binding:"omitempty,oneof=OFFER PRELIMINARY_CONTRACT DEED COMPLETED CANCELLED"`
	NegotiatedPrice *float64   `json:"negotiated_price" binding:"omitempty,gte=0"`
	ContractDate    *time.Time `json:"contract_date" binding:"omitempty"`
	DeedDate        *time.Time `json:"deed_date" binding:"omitempty"`
}

// ToUpdates builds the GORM updates map from the fields actually provided.
func (r UpdateSaleRequest) ToUpdates() map[string]interface{} {
	updates := map[string]interface{}{}
	if r.Status != nil {
		updates["status"] = *r.Status
	}
	if r.NegotiatedPrice != nil {
		updates["negotiated_price"] = *r.NegotiatedPrice
	}
	if r.ContractDate != nil {
		updates["contract_date"] = *r.ContractDate
	}
	if r.DeedDate != nil {
		updates["deed_date"] = *r.DeedDate
	}
	return updates
}
