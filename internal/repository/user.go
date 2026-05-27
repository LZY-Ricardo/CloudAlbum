package repository

import (
	"time"

	"cloudalbum/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePasswordAndBumpVersion 在一个事务内更新密码 hash、自增 TokenVersion、写入 PasswordChangedAt。
func (r *UserRepository) UpdatePasswordAndBumpVersion(userID uint, newHash string, changedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var u model.User
		if err := tx.First(&u, userID).Error; err != nil {
			return err
		}
		u.PasswordHash = newHash
		u.TokenVersion = u.TokenVersion + 1
		u.PasswordChangedAt = &changedAt
		return tx.Save(&u).Error
	})
}
