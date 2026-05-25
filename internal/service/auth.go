package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	cfg       config.AuthConfig
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg}
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateJWT(user *model.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.TokenExpire)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) ParseJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	return s.GenerateJWT(user)
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     strings.TrimSpace(username),
		PasswordHash: string(hashed),
		Role:         "admin",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) EnsureAdmin(username, password string) error {
	_, err := s.userRepo.FindByUsername(strings.TrimSpace(username))
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	_, err = s.Register(username, password)
	return err
}
