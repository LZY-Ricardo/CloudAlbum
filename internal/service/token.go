package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"

	"gorm.io/gorm"
)

var ErrTokenForbidden = errors.New("token forbidden")

type TokenService struct {
	tokenRepo *repository.TokenRepository
}

func NewTokenService(tokenRepo *repository.TokenRepository) *TokenService {
	return &TokenService{tokenRepo: tokenRepo}
}

func (s *TokenService) Create(userID uint, name, scope string) (*model.APIToken, string, error) {
	rawToken, err := generateToken()
	if err != nil {
		return nil, "", err
	}

	hash := sha256.Sum256([]byte(rawToken))
	token := &model.APIToken{
		UserID:    userID,
		Name:      strings.TrimSpace(name),
		TokenHash: hex.EncodeToString(hash[:]),
		Scope:     strings.TrimSpace(scope),
	}

	if err := s.tokenRepo.Create(token); err != nil {
		return nil, "", err
	}

	return token, "ca_" + rawToken, nil
}

func (s *TokenService) Validate(rawToken string) (*model.APIToken, error) {
	normalized := strings.TrimSpace(strings.TrimPrefix(rawToken, "ca_"))
	if normalized == "" {
		return nil, errors.New("invalid token")
	}

	hash := sha256.Sum256([]byte(normalized))
	token, err := s.tokenRepo.FindByHash(hex.EncodeToString(hash[:]))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid token")
		}
		return nil, err
	}

	if err := s.tokenRepo.UpdateLastUsed(token.ID); err != nil {
		return nil, err
	}

	updated, err := s.tokenRepo.FindByID(token.ID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *TokenService) List(userID uint) ([]model.APIToken, error) {
	return s.tokenRepo.ListByUserID(userID)
}

func (s *TokenService) Delete(id uint, userID uint) error {
	token, err := s.tokenRepo.FindByID(id)
	if err != nil {
		return err
	}
	if token.UserID != userID {
		return ErrTokenForbidden
	}

	return s.tokenRepo.Delete(id)
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
