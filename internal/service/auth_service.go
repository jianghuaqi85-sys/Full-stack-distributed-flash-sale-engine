package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type AuthService struct {
	userRepo    repository.UserRepository
	jwtKey      []byte
	expire      int
	redisClient *redis.Client
}

func NewAuthService(userRepo repository.UserRepository, jwtKey string, expire int, redisClient *redis.Client) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtKey:      []byte(jwtKey),
		expire:      expire,
		redisClient: redisClient,
	}
}

type RegisterInput struct {
	Username string
	Password string
	Email    string
}

func (s *AuthService) Register(input RegisterInput) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := db.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Email:    input.Email,
		Role:     "user",
	}

	if err := s.userRepo.Create(&user); err != nil {
		return fmt.Errorf("user already exists")
	}

	return nil
}

type LoginOutput struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
}

func (s *AuthService) Login(username, password string) (*LoginOutput, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Second * time.Duration(s.expire)).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   s.expire,
	}, nil
}

func (s *AuthService) GetProfile(userID uint) (*db.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) UpdateUserRole(userID uint, role string) error {
	if role != "user" && role != "admin" {
		return fmt.Errorf("invalid role: %s", role)
	}

	if err := s.userRepo.UpdateRole(userID, role); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// 主动清除用户缓存，确保下次请求获取最新角色
	if s.redisClient != nil {
		cacheKey := fmt.Sprintf("auth:user:%d", userID)
		s.redisClient.Del(context.Background(), cacheKey)
	}

	return nil
}
