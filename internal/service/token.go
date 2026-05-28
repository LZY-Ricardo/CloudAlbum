package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"

	"gorm.io/gorm"
)

var ErrTokenForbidden = errors.New("token forbidden")

type TokenService struct {
	tokenRepo *repository.TokenRepository
	provider  *config.Provider
}

func NewTokenService(tokenRepo *repository.TokenRepository, provider *config.Provider) *TokenService {
	return &TokenService{tokenRepo: tokenRepo, provider: provider}
}

func (s *TokenService) Create(userID uint, name, scope string, expiresIn *int64) (*model.APIToken, string, error) {
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

	if expiresIn != nil {
		if *expiresIn <= 0 {
			return nil, "", errors.New("invalid expires_in")
		}
		expiresAt := time.Now().Add(time.Duration(*expiresIn) * time.Second)
		token.ExpiresAt = &expiresAt
	} else if s.provider != nil {
		policy := s.provider.Get().Token
		if policy.DefaultExpiresIn > 0 {
			expiresAt := time.Now().Add(policy.DefaultExpiresIn)
			token.ExpiresAt = &expiresAt
		} else if !policy.AllowNoExpiry {
			return nil, "", errors.New("invalid expires_in")
		}
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

	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return nil, errors.New("invalid token")
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
