package dto

type ApplyAsSellerRequest struct {
	Motivation string `json:"motivation" binding:"omitempty,max=500"`
}
