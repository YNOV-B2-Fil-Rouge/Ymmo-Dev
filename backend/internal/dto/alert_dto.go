package dto

// CreateAlertRequest saves a search. All criteria are optional; an empty one
// means "don't filter on this field".
type CreateAlertRequest struct {
	City       string   `json:"city" binding:"omitempty,max=100"`
	CategoryID *uint8   `json:"category_id" binding:"omitempty"`
	MinPrice   *float64 `json:"min_price" binding:"omitempty,gte=0"`
	MaxPrice   *float64 `json:"max_price" binding:"omitempty,gte=0"`
	MinArea    *float64 `json:"min_area" binding:"omitempty,gte=0"`
	MaxEnergy  string   `json:"max_energy" binding:"omitempty,oneof=A B C D E F G"`
}
