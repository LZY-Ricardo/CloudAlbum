package repository

import (
	"fmt"
	"testing"

	"cloudalbum/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestImageRepositoryListExcludesSoftDeletedByDefault(t *testing.T) {
	repo := newTestImageRepository(t)
	userID := createTestUser(t, repo.db, "list-default")
	activeImageID := createTestImage(t, repo.db, userID, "active.jpg", 100)
	deletedImageID := createTestImage(t, repo.db, userID, "deleted.jpg", 200)

	if err := repo.SoftDelete(deletedImageID); err != nil {
		t.Fatalf("soft delete image: %v", err)
	}

	images, total, err := repo.List(ImageListParams{UserID: userID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list images: %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	if images[0].ID != activeImageID {
		t.Fatalf("expected active image %d, got %d", activeImageID, images[0].ID)
	}
}

func TestImageRepositoryListOnlyDeletedReturnsSoftDeletedRows(t *testing.T) {
	repo := newTestImageRepository(t)
	userID := createTestUser(t, repo.db, "list-deleted")
	activeImageID := createTestImage(t, repo.db, userID, "active.jpg", 100)
	deletedImageID := createTestImage(t, repo.db, userID, "deleted.jpg", 200)

	if err := repo.SoftDelete(deletedImageID); err != nil {
		t.Fatalf("soft delete image: %v", err)
	}

	images, total, err := repo.List(ImageListParams{UserID: userID, OnlyDeleted: true, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list deleted images: %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 deleted image, got %d", len(images))
	}
	if images[0].ID != deletedImageID {
		t.Fatalf("expected deleted image %d, got %d", deletedImageID, images[0].ID)
	}
	if images[0].DeletedAt.Time.IsZero() {
		t.Fatal("expected deleted image to include deleted_at timestamp")
	}
	for _, image := range images {
		if image.ID == activeImageID {
			t.Fatalf("expected active image %d to be excluded from deleted listing", activeImageID)
		}
	}
}

func TestImageRepositoryRestoreReturnsImageToDefaultListing(t *testing.T) {
	repo := newTestImageRepository(t)
	userID := createTestUser(t, repo.db, "restore")
	deletedImageID := createTestImage(t, repo.db, userID, "deleted.jpg", 200)

	if err := repo.SoftDelete(deletedImageID); err != nil {
		t.Fatalf("soft delete image: %v", err)
	}

	deletedImages, deletedTotal, err := repo.List(ImageListParams{UserID: userID, OnlyDeleted: true, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list deleted images before restore: %v", err)
	}
	if deletedTotal != 1 || len(deletedImages) != 1 {
		t.Fatalf("expected deleted listing to contain one image before restore, got total=%d len=%d", deletedTotal, len(deletedImages))
	}

	if err := repo.Restore(deletedImageID); err != nil {
		t.Fatalf("restore image: %v", err)
	}

	deletedImages, deletedTotal, err = repo.List(ImageListParams{UserID: userID, OnlyDeleted: true, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list deleted images after restore: %v", err)
	}
	if deletedTotal != 0 || len(deletedImages) != 0 {
		t.Fatalf("expected deleted listing to be empty after restore, got total=%d len=%d", deletedTotal, len(deletedImages))
	}

	images, total, err := repo.List(ImageListParams{UserID: userID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list active images after restore: %v", err)
	}
	if total != 1 || len(images) != 1 {
		t.Fatalf("expected active listing to contain restored image, got total=%d len=%d", total, len(images))
	}
	if images[0].ID != deletedImageID {
		t.Fatalf("expected restored image %d, got %d", deletedImageID, images[0].ID)
	}
}

func TestImageRepositoryListDefaultsInvalidPagination(t *testing.T) {
	repo := newTestImageRepository(t)
	userID := createTestUser(t, repo.db, "pagination")
	for i := 0; i < 25; i++ {
		createTestImage(t, repo.db, userID, testingImageName(i), int64(i+1))
	}

	images, total, err := repo.List(ImageListParams{UserID: userID, Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("list images with default pagination: %v", err)
	}

	if total != 25 {
		t.Fatalf("expected total 25, got %d", total)
	}
	if len(images) != 20 {
		t.Fatalf("expected default page size 20, got %d", len(images))
	}
}

func TestImageRepositoryTotalSizeByUserIDReturnsZeroWhenEmpty(t *testing.T) {
	repo := newTestImageRepository(t)
	userID := createTestUser(t, repo.db, "empty-total")

	total, err := repo.TotalSizeByUserID(userID)
	if err != nil {
		t.Fatalf("total size by user id: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected total size 0, got %d", total)
	}
}

func newTestImageRepository(t *testing.T) *ImageRepository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Album{}, &model.Image{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return NewImageRepository(db)
}

func createTestUser(t *testing.T, db *gorm.DB, username string) uint {
	t.Helper()

	user := &model.User{Username: username, PasswordHash: "hash", Role: "admin"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func createTestImage(t *testing.T, db *gorm.DB, userID uint, originalName string, size int64) uint {
	t.Helper()

	storageKey := originalName + "-key"
	filename := originalName + "-file"
	img := &model.Image{
		UserID:       userID,
		StorageKey:   storageKey,
		Filename:     filename,
		OriginalName: originalName,
		Size:         size,
		MimeType:     "image/jpeg",
		Hash:         storageKey,
	}
	if err := db.Create(img).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return img.ID
}

func testingImageName(i int) string {
	return fmt.Sprintf("image-%02d.jpg", i)
}
