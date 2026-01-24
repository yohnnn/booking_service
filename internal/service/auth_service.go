package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yohnnn/booking_service/internal/models"
	"github.com/yohnnn/booking_service/internal/repository"
)

type AuthService struct {
	userRepo  repository.UserRepository
	secretKey []byte
	tokenTTL  time.Duration
}

func NewAuthService(userRepo repository.UserRepository, secretKey string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: []byte(secretKey),
		tokenTTL:  tokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(bytes),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secretKey)
}

func (s *AuthService) ParseToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, jwt.ErrInvalidKey
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, err
		}

		return userID, nil
	}

	return uuid.Nil, jwt.ErrTokenInvalidClaims
}
