package service

import (
	"errors"
	"fmt"
	"testing"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAlbumServiceUpdateRejectsNonOwner(t *testing.T) {
	db := newTestAlbumServiceDB(t)
	albumRepo := repository.NewAlbumRepository(db)
	imageRepo := repository.NewImageRepository(db)
	svc := NewAlbumService(albumRepo, imageRepo)

	album, err := svc.Create(1, "Travel", "summer")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = svc.Update(album.ID, 2, map[string]interface{}{"name": "Updated", "description": "desc"})
	if !errors.Is(err, ErrImageForbidden) && err == nil {
		t.Fatalf("Update() error = %v, want owner check error", err)
	}
}

func TestAlbumServiceDeleteRejectsNonOwner(t *testing.T) {
	db := newTestAlbumServiceDB(t)
	albumRepo := repository.NewAlbumRepository(db)
	imageRepo := repository.NewImageRepository(db)
	svc := NewAlbumService(albumRepo, imageRepo)

	album, err := svc.Create(1, "Travel", "summer")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	err = svc.Delete(album.ID, 2)
	if !errors.Is(err, ErrImageForbidden) && err == nil {
		t.Fatalf("Delete() error = %v, want owner check error", err)
	}
}

func TestAlbumServiceUpdatePreservesOmittedFields(t *testing.T) {
	db := newTestAlbumServiceDB(t)
	albumRepo := repository.NewAlbumRepository(db)
	imageRepo := repository.NewImageRepository(db)
	svc := NewAlbumService(albumRepo, imageRepo)

	coverID := seedAlbumCoverImage(t, db, 1)
	album, err := svc.Create(1, "Travel", "summer")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	album.CoverImageID = &coverID
	if err := albumRepo.Update(album); err != nil {
		t.Fatalf("Update() seed cover error = %v", err)
	}

	if err := svc.Update(album.ID, 1, map[string]interface{}{"description": "updated"}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := albumRepo.FindByID(album.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Name != "Travel" {
		t.Fatalf("Name = %q, want Travel", updated.Name)
	}
	if updated.Description != "updated" {
		t.Fatalf("Description = %q, want updated", updated.Description)
	}
	if updated.CoverImageID == nil || *updated.CoverImageID != coverID {
		t.Fatalf("CoverImageID changed unexpectedly: %#v", updated.CoverImageID)
	}
}

func seedAlbumCoverImage(t *testing.T, db *gorm.DB, userID uint) uint {
	t.Helper()
	img := &model.Image{
		UserID:       userID,
		StorageKey:   fmt.Sprintf("cover-%d", userID),
		Filename:     "cover.jpg",
		OriginalName: "cover.jpg",
		Size:         100,
		MimeType:     "image/jpeg",
		Hash:         fmt.Sprintf("hash-%d", userID),
	}
	if err := db.Create(img).Error; err != nil {
		t.Fatalf("create cover image: %v", err)
	}
	return img.ID
}

func newTestAlbumServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Album{}, &model.Image{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	users := []model.User{
		{Username: "owner", PasswordHash: "hash", Role: "admin"},
		{Username: "other", PasswordHash: "hash", Role: "admin"},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user %d: %v", i, err)
		}
	}

	return db
}
