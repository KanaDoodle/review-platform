package service

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"review-platform/internal/model"
	"review-platform/internal/repository"
	jwtutil "review-platform/pkg/jwt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo     *repository.UserRepository
	rdb          *redis.Client
	jwtSecret    string
	jwtExpireHrs int
}

func NewAuthService(
	userRepo *repository.UserRepository,
	rdb *redis.Client,
	jwtSecret string,
	jwtExpireHrs int,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		rdb:          rdb,
		jwtSecret:    jwtSecret,
		jwtExpireHrs: jwtExpireHrs,
	}
}

func (s *AuthService) SendCode(phone string) error {
	if !isValidPhone(phone) {
		return ErrInvalidPhone
	}

	// MVP：先固定验证码，后面可改成随机数 + 短信服务
	code := "123456"

	key := loginCodeKey(phone)
	return s.rdb.Set(context.Background(), key, code, 5*time.Minute).Err()
}

func (s *AuthService) Login(phone, code string) (string, error) {
	if !isValidPhone(phone) {
		return "", ErrInvalidPhone
	}

	key := loginCodeKey(phone)
	storedCode, err := s.rdb.Get(context.Background(), key).Result()
	if err != nil {
		return "", ErrInvalidCode
	}

	if storedCode != code {
		return "", ErrInvalidCode
	}

	user, err := s.userRepo.GetByPhone(phone)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user = &model.User{
				Phone:    phone,
				Nickname: generateNickname(),
			}
			if err := s.userRepo.Create(user); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	token, err := jwtutil.GenerateToken(s.jwtSecret, s.jwtExpireHrs, user.ID, user.Phone)
	if err != nil {
		return "", err
	}

	return token, nil
}

func loginCodeKey(phone string) string {
	return fmt.Sprintf("login:code:%s", phone)
}

func isValidPhone(phone string) bool {
	// 简化版中国大陆手机号校验
	re := regexp.MustCompile(`^1\d{10}$`)
	return re.MatchString(phone)
}

func generateNickname() string {
	return fmt.Sprintf("user_%06d", rand.Intn(1000000))
}