package dto

type CreatePropertyRequest struct {
	Title       string  `json:"title" binding:"required,min=3,max=150"`
	Description string  `json:"description" binding:"omitempty"`
	CategoryID  uint8   `json:"category_id" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Area        float64 `json:"area" binding:"required,gt=0"`
	Rooms       *uint8  `json:"rooms" binding:"omitempty"`
	Bedrooms    *uint8  `json:"bedrooms" binding:"omitempty"`
	Bathrooms   *uint8  `json:"bathrooms" binding:"omitempty"`
	Floor       *int8   `json:"floor" binding:"omitempty"`
	BuildYear   *uint16 `json:"build_year" binding:"omitempty"`
	EnergyRating string `json:"energy_rating" binding:"omitempty,oneof=A B C D E F G"`
	GhgRating   string  `json:"ghg_rating" binding:"omitempty,oneof=A B C D E F G"`
	Address     string  `json:"address" binding:"omitempty,max=255"`
	City        string  `json:"city" binding:"required,max=100"`
	PostalCode  string  `json:"postal_code" binding:"required,max=10"`
	Latitude    *float64 `json:"latitude" binding:"omitempty"`
	Longitude   *float64 `json:"longitude" binding:"omitempty"`
	IsExclusive bool    `json:"is_exclusive"`
	AgencyID    uint16  `json:"agency_id" binding:"required"`
}

type UpdatePropertyRequest struct {
	Title       *string  `json:"title" binding:"omitempty,min=3,max=150"`
	Description *string  `json:"description" binding:"omitempty"`
	CategoryID  *uint8   `json:"category_id" binding:"omitempty"`
	Status      *string  `json:"status" binding:"omitempty,oneof=DRAFT PENDING_REVIEW AVAILABLE UNDER_OFFER SOLD WITHDRAWN"`
	Price       *float64 `json:"price" binding:"omitempty,gt=0"`
	Area        *float64 `json:"area" binding:"omitempty,gt=0"`
	Rooms       *uint8   `json:"rooms" binding:"omitempty"`
	Bedrooms    *uint8   `json:"bedrooms" binding:"omitempty"`
	Bathrooms   *uint8   `json:"bathrooms" binding:"omitempty"`
	Floor       *int8    `json:"floor" binding:"omitempty"`
	BuildYear   *uint16  `json:"build_year" binding:"omitempty"`
	EnergyRating *string `json:"energy_rating" binding:"omitempty,oneof=A B C D E F G"`
	GhgRating   *string  `json:"ghg_rating" binding:"omitempty,oneof=A B C D E F G"`
	Address     *string  `json:"address" binding:"omitempty,max=255"`
	City        *string  `json:"city" binding:"omitempty,max=100"`
	PostalCode  *string  `json:"postal_code" binding:"omitempty,max=10"`
	IsExclusive *bool    `json:"is_exclusive" binding:"omitempty"`
}

func (r UpdatePropertyRequest) ToUpdates() map[string]interface{} {
	updates := map[string]interface{}{}
	if r.Title != nil {
		updates["title"] = *r.Title
	}
	if r.Description != nil {
		updates["description"] = *r.Description
	}
	if r.CategoryID != nil {
		updates["category_id"] = *r.CategoryID
	}
	if r.Status != nil {
		updates["status"] = *r.Status
	}
	if r.Price != nil {
		updates["price"] = *r.Price
	}
	if r.Area != nil {
		updates["area"] = *r.Area
	}
	if r.Rooms != nil {
		updates["rooms"] = *r.Rooms
	}
	if r.Bedrooms != nil {
		updates["bedrooms"] = *r.Bedrooms
	}
	if r.Bathrooms != nil {
		updates["bathrooms"] = *r.Bathrooms
	}
	if r.Floor != nil {
		updates["floor"] = *r.Floor
	}
	if r.BuildYear != nil {
		updates["build_year"] = *r.BuildYear
	}
	if r.EnergyRating != nil {
		updates["energy_rating"] = *r.EnergyRating
	}
	if r.GhgRating != nil {
		updates["ghg_rating"] = *r.GhgRating
	}
	if r.Address != nil {
		updates["address"] = *r.Address
	}
	if r.City != nil {
		updates["city"] = *r.City
	}
	if r.PostalCode != nil {
		updates["postal_code"] = *r.PostalCode
	}
	if r.IsExclusive != nil {
		updates["is_exclusive"] = *r.IsExclusive
	}
	return updates
}

type PropertySearchQuery struct {
	City       string  `form:"city"`
	CategoryID uint8   `form:"category_id"`
	Sector     string  `form:"sector"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	MinArea    float64 `form:"min_area"`
	MaxArea    float64 `form:"max_area"`
	MaxEnergy  string  `form:"max_energy"`
	Status     string  `form:"status"`
	Page       int     `form:"page"`
	PageSize   int     `form:"page_size"`
	Sort       string  `form:"sort"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
