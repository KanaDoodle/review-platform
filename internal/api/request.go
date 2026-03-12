package api

type CreateReviewRequest struct {
	UserID  int64  `json:"user_id"`
	ShopID  int64  `json:"shop_id"`
	Content string `json:"content"`
}

