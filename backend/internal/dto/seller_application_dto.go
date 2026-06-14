package dto

// ApplyAsSellerRequest is a buyer's request to become a seller. The motivation
// is optional free text shown to the reviewing agent.
type ApplyAsSellerRequest struct {
	Motivation string `json:"motivation" binding:"omitempty,max=500"`
}
