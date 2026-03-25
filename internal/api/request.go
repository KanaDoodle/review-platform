package api

type CreateReviewRequest struct {
	ShopID  int64  `json:"shop_id"`
	Content string `json:"content"`
}

type SendCodeRequest struct {
	Phone string `json:"phone"`
}

type LoginRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type NearbyShopItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	CategoryID  int64   `json:"category_id"`
	Address     string  `json:"address"`
	Lng         float64 `json:"lng"`
	Lat         float64 `json:"lat"`
	Score       float64 `json:"score"`
	AvgPrice    int     `json:"avg_price"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
}