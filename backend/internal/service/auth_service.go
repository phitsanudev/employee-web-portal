package service

import (
	"errors"
	"time"

	"employee-portal/backend/internal/config"
	"employee-portal/backend/internal/domain"
	"employee-portal/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg  config.Config
	repo *repository.ProfileRepository
}

type LoginResult struct {
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expiresAt"`
	User      UserDTO `json:"user"`
}

type Claims struct {
	UserID uint   `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg config.Config, repo *repository.ProfileRepository) *AuthService {
	return &AuthService{cfg: cfg, repo: repo}
}

func (s *AuthService) Login(email, password string) (*LoginResult, error) {
	if email == "" || password == "" {
		return nil, domain.NewAppError(domain.ErrValidation, "VALIDATION_ERROR", "Email and password are required")
	}
	user, err := s.repo.FindUserByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Invalid email or password")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Invalid email or password")
	}

	expiresAt := time.Now().Add(time.Duration(s.cfg.JWTExpiresHours) * time.Hour)
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Email,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		Token: token, ExpiresAt: expiresAt.Format(time.RFC3339),
		User: UserDTO{ID: user.ID, Email: user.Email, Role: user.Role},
	}, nil
}

func (s *AuthService) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Invalid token claims")
	}
	return claims, nil
}
