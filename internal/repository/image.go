package repository

import (
	"cloudalbum/internal/model"
	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(img *model.Image) error {
	return r.db.Create(img).Error
}

func (r *ImageRepository) FindByID(id uint) (*model.Image, error) {
	var img model.Image
	if err := r.db.First(&img, id).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *ImageRepository) FindByHash(hash string) (*model.Image, error) {
	var img model.Image
	if err := r.db.Where("hash = ?", hash).First(&img).Error; err != nil {
		return nil, err
	}
	return &img, nil
}

type ImageListParams struct {
	UserID      uint
	AlbumID     *uint
	Page        int
	PageSize    int
	Keyword     string
	OnlyDeleted bool
}

func (r *ImageRepository) List(params ImageListParams) ([]model.Image, int64, error) {
	var images []model.Image
	var total int64

	query := r.db.Model(&model.Image{})
	if params.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}
	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.AlbumID != nil {
		query = query.Where("album_id = ?", *params.AlbumID)
	}
	if params.Keyword != "" {
		query = query.Where("original_name LIKE ?", "%"+params.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&images).Error; err != nil {
		return nil, 0, err
	}
	return images, total, nil
}

func (r *ImageRepository) Update(img *model.Image) error {
	return r.db.Save(img).Error
}

func (r *ImageRepository) SoftDelete(id uint) error {
	return r.db.Delete(&model.Image{}, id).Error
}

func (r *ImageRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&model.Image{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *ImageRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&model.Image{}, id).Error
}

func (r *ImageRepository) BatchUpdate(ids []uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Image{}).Where("id IN ?", ids).Updates(updates).Error
}

func (r *ImageRepository) BatchDelete(ids []uint) error {
	return r.db.Where("id IN ?", ids).Delete(&model.Image{}).Error
}

func (r *ImageRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Image{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *ImageRepository) TotalSizeByUserID(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&model.Image{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&total).Error
	return total, err
}
