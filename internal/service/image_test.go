package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloudalbum/internal/config"
	imgpkg "cloudalbum/internal/image"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestImageServiceUploadDeduplicatesByHash(t *testing.T) {
	db := newTestImageServiceDB(t)
	imageRepo := repository.NewImageRepository(db)
	store, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}

	svc := NewImageService(
		imageRepo,
		store,
		imgpkg.NewProcessor(testImageConfig()),
		testImageConfig(),
		"http://localhost:8080",
	)

	fileA := newTestFileHeader(t, "sample.jpg", testJPEGBytes(t))
	createdA, err := svc.Upload(1, fileA, nil)
	if err != nil {
		t.Fatalf("Upload() first error = %v", err)
	}

	fileB := newTestFileHeader(t, "renamed.jpg", testJPEGBytes(t))
	createdB, err := svc.Upload(1, fileB, nil)
	if err != nil {
		t.Fatalf("Upload() duplicate error = %v", err)
	}

	if createdA.ID == createdB.ID {
		t.Fatalf("expected duplicate upload to create a new row, got same id %d", createdA.ID)
	}
	if createdA.StorageKey != createdB.StorageKey {
		t.Fatalf("expected duplicate upload to reuse storage key, got %q and %q", createdA.StorageKey, createdB.StorageKey)
	}
	if createdB.OriginalName != "renamed.jpg" {
		t.Fatalf("duplicate row original name = %q, want renamed.jpg", createdB.OriginalName)
	}

	count, err := imageRepo.CountByUserID(1)
	if err != nil {
		t.Fatalf("CountByUserID() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 image rows after duplicate upload, got %d", count)
	}

	exists, err := store.Exists(nil, createdA.StorageKey)
	if err != nil {
		t.Fatalf("Exists() original error = %v", err)
	}
	if !exists {
		t.Fatalf("expected stored original at %q", createdA.StorageKey)
	}

	thumbKey := imgpkg.NewProcessor(testImageConfig()).GenerateThumbnailKey(createdA.StorageKey, "thumb")
	exists, err = store.Exists(nil, thumbKey)
	if err != nil {
		t.Fatalf("Exists() thumbnail error = %v", err)
	}
	if !exists {
		t.Fatalf("expected stored thumbnail at %q", thumbKey)
	}
}

func TestImageServiceUploadRejectsDisallowedExtension(t *testing.T) {
	db := newTestImageServiceDB(t)
	imageRepo := repository.NewImageRepository(db)
	store, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}

	svc := NewImageService(
		imageRepo,
		store,
		imgpkg.NewProcessor(testImageConfig()),
		testImageConfig(),
		"http://localhost:8080",
	)

	file := newTestFileHeader(t, "bad.txt", testJPEGBytes(t))
	_, err = svc.Upload(1, file, nil)
	if err == nil {
		t.Fatal("Upload() error = nil, want disallowed extension error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("Upload() error = %v, want not allowed", err)
	}
}

func TestImageServiceURLsBuildExpectedFormats(t *testing.T) {
	svc := NewImageService(nil, nil, nil, testImageConfig(), "http://localhost:8080")
	img := &model.Image{OriginalName: "hello.jpg", StorageKey: "1/abcd/hello.jpg"}

	urls := svc.URLs(img)
	if urls["url"] != "http://localhost:8080/i/1/abcd/hello.jpg" {
		t.Fatalf("url = %q", urls["url"])
	}
	if urls["markdown"] != "![hello.jpg](http://localhost:8080/i/1/abcd/hello.jpg)" {
		t.Fatalf("markdown = %q", urls["markdown"])
	}
	if !strings.Contains(urls["html"], `<img src="http://localhost:8080/i/1/abcd/hello.jpg" alt="hello.jpg">`) {
		t.Fatalf("html = %q", urls["html"])
	}
	if urls["bbcode"] != "[img]http://localhost:8080/i/1/abcd/hello.jpg[/img]" {
		t.Fatalf("bbcode = %q", urls["bbcode"])
	}
}

func TestImageServiceUpdatePreservesAlbumWhenAlbumIDOmitted(t *testing.T) {
	db := newTestImageServiceDB(t)
	imageRepo := repository.NewImageRepository(db)
	svc := NewImageService(imageRepo, nil, nil, testImageConfig(), "http://localhost:8080")

	albumID := uint(7)
	img := &model.Image{
		UserID:       1,
		StorageKey:   "1/preserve/test.jpg",
		Filename:     "test.jpg",
		OriginalName: "test.jpg",
		Size:         123,
		MimeType:     "image/jpeg",
		Hash:         "preserve-hash",
		AlbumID:      &albumID,
	}
	if err := db.Create(img).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	if err := svc.Update(img.ID, 1, map[string]interface{}{"original_name": "renamed.jpg"}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, err := imageRepo.FindByID(img.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.OriginalName != "renamed.jpg" {
		t.Fatalf("OriginalName = %q, want renamed.jpg", updated.OriginalName)
	}
	if updated.AlbumID == nil || *updated.AlbumID != albumID {
		t.Fatalf("AlbumID changed unexpectedly: %#v", updated.AlbumID)
	}
}

func newTestImageServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Album{}, &model.Image{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	user := &model.User{Username: "image-user", PasswordHash: "hash", Role: "admin"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return db
}

func newTestFileHeader(t *testing.T, filename string, data []byte) *multipart.FileHeader {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(len(body.Bytes())))
	if err != nil {
		t.Fatalf("ReadForm() error = %v", err)
	}
	files := form.File["files"]
	if len(files) != 1 {
		t.Fatalf("expected one file header, got %d", len(files))
	}
	return files[0]
}

func testImageConfig() config.ImageConfig {
	return config.ImageConfig{
		MaxSize:      5 << 20,
		AllowedTypes: []string{"jpg", "jpeg", "png", "gif", "webp", "bmp"},
		AutoConvert:  "jpeg",
		Quality:      85,
		Thumbnails: []config.ThumbnailSize{
			{Name: "thumb", Width: 50, Height: 50},
		},
	}
}

func testJPEGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode() error = %v", err)
	}
	return buf.Bytes()
}
