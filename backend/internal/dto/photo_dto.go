package dto

type AddPhotoRequest struct {
	URL       string `json:"url" binding:"required,url,max=255"`
	IsPrimary bool   `json:"is_primary"`
}
