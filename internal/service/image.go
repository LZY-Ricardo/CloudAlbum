package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloudalbum/internal/config"
	imgpkg "cloudalbum/internal/image"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrImageForbidden = errors.New("image forbidden")

type ImageService struct {
	imageRepo *repository.ImageRepository
	store     storage.Storage
	processor *imgpkg.Processor
	provider  *config.Provider
}

func NewImageService(imageRepo *repository.ImageRepository, store storage.Storage, processor *imgpkg.Processor, provider *config.Provider) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		store:     store,
		processor: processor,
		provider:  provider,
	}
}

func (s *ImageService) imageCfg() config.ImageConfig {
	return s.provider.Get().Image
}

func (s *ImageService) baseURL() string {
	return strings.TrimRight(s.provider.Get().Server.BaseURL, "/")
}

func (s *ImageService) Upload(userID uint, file *multipart.FileHeader, albumID *uint) (*model.Image, error) {
	if file == nil {
		return nil, errors.New("file required")
	}
	if file.Size > s.imageCfg().MaxSize {
		return nil, fmt.Errorf("file too large")
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), ".")
	if !s.isAllowedType(ext) {
		return nil, fmt.Errorf("file type %s not allowed", ext)
	}

	opened, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer opened.Close()

	data, err := io.ReadAll(opened)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}

	mimeType := imgpkg.DetectImageType(data)
	if mimeType == "" {
		return nil, fmt.Errorf("unsupported image type")
	}
	if !s.isAllowedType(mimeTypeToExt(mimeType)) {
		return nil, fmt.Errorf("file type %s not allowed", mimeTypeToExt(mimeType))
	}

	return s.storeProcessedImage(userID, file.Filename, data, mimeType, albumID)
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func (s *ImageService) UploadFromURL(userID uint, imageURL string, albumID *uint) (*model.Image, error) {
	resp, err := httpClient.Get(strings.TrimSpace(imageURL))
	if err != nil {
		return nil, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch url: unexpected status %d", resp.StatusCode)
	}

	maxSize := s.imageCfg().MaxSize
	limited := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read remote image: %w", err)
	}
	if int64(len(data)) > maxSize {
		return nil, fmt.Errorf("file too large")
	}

	mimeType := imgpkg.DetectImageType(data)
	if mimeType == "" {
		mimeType = strings.TrimSpace(resp.Header.Get("Content-Type"))
		if idx := strings.Index(mimeType, ";"); idx >= 0 {
			mimeType = strings.TrimSpace(mimeType[:idx])
		}
	}
	if mimeType == "" {
		return nil, fmt.Errorf("unsupported image type")
	}

	ext := mimeTypeToExt(mimeType)
	if !s.isAllowedType(ext) {
		return nil, fmt.Errorf("file type %s not allowed", ext)
	}

	filename := remoteFilename(imageURL, ext)
	return s.storeProcessedImage(userID, filename, data, mimeType, albumID)
}

func (s *ImageService) Get(id uint) (*model.Image, error) {
	return s.imageRepo.FindByID(id)
}

func (s *ImageService) List(params repository.ImageListParams) ([]model.Image, int64, error) {
	return s.imageRepo.List(params)
}

func (s *ImageService) Update(id uint, userID uint, updates map[string]interface{}) error {
	img, err := s.imageRepo.FindByID(id)
	if err != nil {
		return err
	}
	if img.UserID != userID {
		return ErrImageForbidden
	}

	if rawName, ok := updates["original_name"]; ok {
		name, ok := rawName.(string)
		if !ok {
			return fmt.Errorf("original_name must be a string")
		}
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			img.OriginalName = trimmed
		}
	}
	if albumValue, ok := updates["album_id"]; ok {
		switch v := albumValue.(type) {
		case nil:
			img.AlbumID = nil
		case float64:
			converted := uint(v)
			img.AlbumID = &converted
		case uint:
			img.AlbumID = &v
		case int:
			converted := uint(v)
			img.AlbumID = &converted
		default:
			return fmt.Errorf("album_id must be a number or null")
		}
	}

	return s.imageRepo.Update(img)
}

func (s *ImageService) Delete(id uint, userID uint) error {
	img, err := s.imageRepo.FindByID(id)
	if err != nil {
		return err
	}
	if img.UserID != userID {
		return ErrImageForbidden
	}
	return s.imageRepo.SoftDelete(id)
}

func (s *ImageService) Restore(id uint, userID uint) error {
	img, err := s.imageRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if img.UserID != userID {
		return ErrImageForbidden
	}
	return s.imageRepo.Restore(id)
}

func (s *ImageService) HardDelete(id uint, userID uint) error {
	img, err := s.imageRepo.FindByIDUnscoped(id)
	if err != nil {
		return err
	}
	if img.UserID != userID {
		return ErrImageForbidden
	}
	return s.imageRepo.HardDelete(id)
}

func (s *ImageService) BatchOperation(ids []uint, userID uint, action string, albumID *uint) error {
	if len(ids) == 0 {
		return nil
	}
	owned, err := s.imageRepo.CountOwnedByUser(ids, userID)
	if err != nil {
		return err
	}
	if owned != int64(len(ids)) {
		return ErrImageForbidden
	}

	switch strings.ToLower(strings.TrimSpace(action)) {
	case "delete":
		return s.imageRepo.BatchDelete(ids)
	case "move":
		return s.imageRepo.BatchUpdate(ids, map[string]interface{}{"album_id": albumID})
	default:
		return fmt.Errorf("unknown batch action: %s", action)
	}
}

func (s *ImageService) Stats(userID uint) (count int64, totalSize int64, err error) {
	count, err = s.imageRepo.CountByUserID(userID)
	if err != nil {
		return 0, 0, err
	}
	totalSize, err = s.imageRepo.TotalSizeByUserID(userID)
	if err != nil {
		return 0, 0, err
	}
	return count, totalSize, nil
}

func (s *ImageService) URLs(img *model.Image) map[string]string {
	url := s.baseURL() + "/i/" + strings.TrimLeft(img.StorageKey, "/")
	return map[string]string{
		"url":      url,
		"markdown": fmt.Sprintf("![%s](%s)", img.OriginalName, url),
		"html":     fmt.Sprintf(`<img src="%s" alt="%s">`, url, img.OriginalName),
		"bbcode":   fmt.Sprintf("[img]%s[/img]", url),
	}
}

func (s *ImageService) isAllowedType(ext string) bool {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	for _, allowed := range s.imageCfg().AllowedTypes {
		if normalized == strings.TrimPrefix(strings.ToLower(strings.TrimSpace(allowed)), ".") {
			return true
		}
	}
	return false
}

func mimeTypeToExt(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	case "image/svg+xml":
		return "svg"
	default:
		return ""
	}
}

func (s *ImageService) storeProcessedImage(userID uint, originalName string, data []byte, mimeType string, albumID *uint) (*model.Image, error) {
	result, err := s.processor.Process(data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("process image: %w", err)
	}

	existing, err := s.imageRepo.FindByHash(result.Hash)
	if err == nil && existing != nil {
		dup := &model.Image{
			UserID:       userID,
			StorageKey:   existing.StorageKey,
			Filename:     existing.Filename,
			OriginalName: originalName,
			Size:         existing.Size,
			MimeType:     existing.MimeType,
			Width:        existing.Width,
			Height:       existing.Height,
			Hash:         existing.Hash,
			AlbumID:      albumID,
		}
		if err := s.imageRepo.Create(dup); err != nil {
			return nil, err
		}
		return dup, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	key := buildStorageKey(userID, originalName)
	storedOriginal := data
	if len(result.OriginalData) > 0 {
		storedOriginal = result.OriginalData
	}
	ctx := context.Background()
	if err := s.store.Save(ctx, key, bytes.NewReader(storedOriginal)); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}
	for sizeName, thumbData := range result.Thumbnails {
		thumbKey := s.processor.GenerateThumbnailKey(key, sizeName)
		if err := s.store.Save(ctx, thumbKey, bytes.NewReader(thumbData)); err != nil {
			return nil, fmt.Errorf("save thumbnail %s: %w", sizeName, err)
		}
	}

	img := &model.Image{
		UserID:       userID,
		StorageKey:   key,
		Filename:     filepath.Base(key),
		OriginalName: originalName,
		Size:         result.Size,
		MimeType:     result.MimeType,
		Width:        result.Width,
		Height:       result.Height,
		Hash:         result.Hash,
		AlbumID:      albumID,
	}
	if err := s.imageRepo.Create(img); err != nil {
		return nil, err
	}
	return img, nil
}

func buildStorageKey(userID uint, filename string) string {
	safeName := strings.ReplaceAll(filepath.Base(strings.TrimSpace(filename)), " ", "_")
	if safeName == "." || safeName == "" || safeName == string(filepath.Separator) {
		safeName = "image"
	}
	prefix := strings.ReplaceAll(uuid.New().String(), "-", "")[:8]
	return strconv.FormatUint(uint64(userID), 10) + "/" + prefix + "/" + safeName
}

func remoteFilename(imageURL, ext string) string {
	trimmed := strings.TrimSpace(imageURL)
	base := filepath.Base(trimmed)
	if idx := strings.Index(base, "?"); idx >= 0 {
		base = base[:idx]
	}
	if base == "." || base == "/" || base == "" {
		return "url_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8] + "." + ext
	}
	if filepath.Ext(base) == "" && ext != "" {
		return base + "." + ext
	}
	return base
}
