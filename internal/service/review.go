package service

import(
	"strings"
	"review-platform/internal/model"
	"review-platform/internal/repository"
)

type ReviewService struct {
	reviewRepo *repository.ReviewRepository
	shopRepo  *repository.ShopRepository
}

func NewReviewService(reviewRepo *repository.ReviewRepository, shopRepo *repository.ShopRepository) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		shopRepo:  shopRepo,
	}
}

func (s *ReviewService) ListByShopID(shopID int64) ([]model.Review, error){
	_, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		
		return nil, ErrShopNotFound
	}

	return s.reviewRepo.ListByShopID(shopID)
}

func (s *ReviewService) Create(userID, shopID int64, content string) error {
	content = strings.TrimSpace(content)

	if userID <= 0 || shopID <= 0 || content == "" {
		return ErrInvalidReviewData
	}

	if len([]rune(content))>500 {
		return ErrInvalidReviewData
	}

	_, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return ErrShopNotFound
	}

	review := &model.Review{
		UserID:  userID,
		ShopID:  shopID,
		Content: content,
		LikeCount: 0,
	}

	return s.reviewRepo.Create(review)
}