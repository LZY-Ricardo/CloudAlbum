package repository

import (
	"time"

	"cloudalbum/internal/model"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(token *model.APIToken) error {
	return r.db.Create(token).Error
}

func (r *TokenRepository) FindByID(id uint) (*model.APIToken, error) {
	var t model.APIToken
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) FindByHash(tokenHash string) (*model.APIToken, error) {
	var t model.APIToken
	if err := r.db.Where("token_hash = ?", tokenHash).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) ListByUserID(userID uint) ([]model.APIToken, error) {
	var tokens []model.APIToken
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *TokenRepository) Delete(id uint) error {
	return r.db.Delete(&model.APIToken{}, id).Error
}

func (r *TokenRepository) UpdateLastUsed(id uint) error {
	return r.db.Model(&model.APIToken{}).Where("id = ?", id).UpdateColumn("last_used_at", time.Now()).Error
}
