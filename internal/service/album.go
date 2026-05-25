package service

import (
	"errors"
	"fmt"
	"strings"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"gorm.io/gorm"
)

type AlbumService struct {
	albumRepo *repository.AlbumRepository
	imageRepo *repository.ImageRepository
}

func NewAlbumService(albumRepo *repository.AlbumRepository, imageRepo *repository.ImageRepository) *AlbumService {
	return &AlbumService{albumRepo: albumRepo, imageRepo: imageRepo}
}

func (s *AlbumService) Create(userID uint, name, description string) (*model.Album, error) {
	album := &model.Album{
		UserID:      userID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
	}
	if err := s.albumRepo.Create(album); err != nil {
		return nil, err
	}
	return album, nil
}

func (s *AlbumService) Get(id uint) (*model.Album, error) {
	return s.albumRepo.FindByID(id)
}

func (s *AlbumService) List(userID uint) ([]model.Album, error) {
	return s.albumRepo.ListByUserID(userID)
}

func (s *AlbumService) Update(id uint, userID uint, updates map[string]interface{}) error {
	album, err := s.albumRepo.FindByID(id)
	if err != nil {
		return err
	}
	if album.UserID != userID {
		return ErrImageForbidden
	}

	if rawName, ok := updates["name"]; ok {
		name, ok := rawName.(string)
		if !ok {
			return fmt.Errorf("name must be a string")
		}
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			album.Name = trimmed
		}
	}
	if rawDescription, ok := updates["description"]; ok {
		description, ok := rawDescription.(string)
		if !ok {
			return fmt.Errorf("description must be a string")
		}
		album.Description = strings.TrimSpace(description)
	}
	if rawCover, ok := updates["cover_image_id"]; ok {
		switch v := rawCover.(type) {
		case nil:
			album.CoverImageID = nil
		case float64:
			coverID := uint(v)
			img, err := s.imageRepo.FindByID(coverID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("cover image not found")
				}
				return err
			}
			if img.UserID != userID {
				return ErrImageForbidden
			}
			album.CoverImageID = &coverID
		default:
			return fmt.Errorf("cover_image_id must be a number or null")
		}
	}

	return s.albumRepo.Update(album)
}

func (s *AlbumService) Delete(id uint, userID uint) error {
	album, err := s.albumRepo.FindByID(id)
	if err != nil {
		return err
	}
	if album.UserID != userID {
		return ErrImageForbidden
	}
	if err := s.imageRepo.ClearAlbum(id, userID); err != nil {
		return err
	}
	return s.albumRepo.Delete(id)
}
