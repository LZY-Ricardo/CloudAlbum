package repository

import (
	"cloudalbum/internal/model"
	"gorm.io/gorm"
)

type AlbumRepository struct {
	db *gorm.DB
}

func NewAlbumRepository(db *gorm.DB) *AlbumRepository {
	return &AlbumRepository{db: db}
}

func (r *AlbumRepository) Create(album *model.Album) error {
	return r.db.Create(album).Error
}

func (r *AlbumRepository) FindByID(id uint) (*model.Album, error) {
	var album model.Album
	if err := r.db.First(&album, id).Error; err != nil {
		return nil, err
	}
	return &album, nil
}

func (r *AlbumRepository) ListByUserID(userID uint) ([]model.Album, error) {
	var albums []model.Album
	if err := r.db.Where("user_id = ?", userID).Order("sort_order ASC, created_at DESC").Find(&albums).Error; err != nil {
		return nil, err
	}
	return albums, nil
}

func (r *AlbumRepository) Update(album *model.Album) error {
	return r.db.Save(album).Error
}

func (r *AlbumRepository) Delete(id uint) error {
	return r.db.Delete(&model.Album{}, id).Error
}

func (r *AlbumRepository) CountImages(albumID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Image{}).Where("album_id = ?", albumID).Count(&count).Error
	return count, err
}
