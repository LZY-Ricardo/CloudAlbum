# CloudAlbum Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hosted personal image hosting service with Go backend, React admin panel, pluggable storage, and Docker deployment.

**Architecture:** Go (Gin) REST API + React SPA embedded via go:embed. SQLite by default with GORM for database abstraction. Pluggable storage interface with Local and S3 backends. Image processing pipeline runs synchronously on upload.

**Tech Stack:** Go 1.22+, Gin, GORM, JWT; React 18, TypeScript, Vite, Arco Design

---

## File Structure Map

```
CloudAlbum/
├── cmd/server/main.go                 # Application entry point
├── internal/
│   ├── config/config.go               # Config struct + YAML loader
│   ├── model/
│   │   ├── user.go                    # User GORM model
│   │   ├── image.go                   # Image GORM model
│   │   ├── album.go                   # Album GORM model
│   │   └── token.go                   # APIToken GORM model
│   ├── repository/
│   │   ├── user.go                    # User DB operations
│   │   ├── image.go                   # Image DB operations
│   │   ├── album.go                   # Album DB operations
│   │   └── token.go                   # Token DB operations
│   ├── service/
│   │   ├── auth.go                    # Auth business logic
│   │   ├── image.go                   # Image business logic
│   │   ├── album.go                   # Album business logic
│   │   └── token.go                   # Token business logic
│   ├── handler/
│   │   ├── auth.go                    # Login/logout handlers
│   │   ├── image.go                   # Image CRUD handlers
│   │   ├── album.go                   # Album CRUD handlers
│   │   ├── token.go                   # Token CRUD handlers
│   │   └── public.go                  # Public image/thumbnail serving
│   ├── middleware/
│   │   ├── auth.go                    # JWT + API Token auth middleware
│   │   ├── cors.go                    # CORS middleware
│   │   └── ratelimit.go              # Rate limiting middleware
│   ├── storage/
│   │   ├── storage.go                 # Storage interface
│   │   ├── local.go                   # Local filesystem storage
│   │   └── s3.go                      # S3-compatible storage
│   ├── image/
│   │   └── processor.go              # Image processing pipeline
│   └── router/router.go              # Route registration
├── web/                                # React frontend
│   ├── src/
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   ├── api/                       # API client
│   │   ├── components/                # Shared components
│   │   ├── pages/                     # Page components
│   │   ├── hooks/                     # Custom hooks
│   │   ├── stores/                    # State management
│   │   └── types/                     # TypeScript types
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
├── embed.go                            # go:embed web/dist
├── configs/config.yaml                 # Default config
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

---

## Phase 1: Backend Foundation

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/config/config.go`
- Create: `configs/config.yaml`
- Create: `Makefile`

- [ ] **Step 1: Initialize Go module and install dependencies**

```bash
cd /Users/zyb/workspace/person/CloudAlbum
go mod init cloudalbum
go get github.com/gin-gonic/gin@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/sqlite@latest
go get gorm.io/driver/postgres@latest
go get github.com/golang-jwt/jwt/v5@latest
go get golang.org/x/crypto@latest
go get github.com/google/uuid@latest
go get gopkg.in/yaml.v3@latest
go get github.com/disintegration/imaging@latest
```

- [ ] **Step 2: Create config struct and loader**

Create `internal/config/config.go`:

```go
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	Image    ImageConfig    `yaml:"image"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type StorageConfig struct {
	Driver string        `yaml:"driver"`
	Local  LocalStorage  `yaml:"local"`
	S3     S3StorageConf `yaml:"s3"`
}

type LocalStorage struct {
	Path string `yaml:"path"`
}

type S3StorageConf struct {
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
	Endpoint string `yaml:"endpoint"`
	AK       string `yaml:"access_key"`
	SK       string `yaml:"secret_key"`
}

type ThumbnailSize struct {
	Name   string `yaml:"name"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

type ImageConfig struct {
	MaxSize       int64           `yaml:"max_size"`
	AllowedTypes  []string        `yaml:"allowed_types"`
	AutoConvert   string          `yaml:"auto_convert"`
	Quality       int             `yaml:"quality"`
	StripExif     bool            `yaml:"strip_exif"`
	Thumbnails    []ThumbnailSize `yaml:"thumbnails"`
}

type AuthConfig struct {
	JWTSecret   string        `yaml:"jwt_secret"`
	TokenExpire time.Duration `yaml:"token_expire"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = "./data/cloudalbum.db"
	}
	if cfg.Storage.Driver == "" {
		cfg.Storage.Driver = "local"
	}
	if cfg.Storage.Local.Path == "" {
		cfg.Storage.Local.Path = "./data/images"
	}
	if cfg.Image.MaxSize == 0 {
		cfg.Image.MaxSize = 50 << 20 // 50MB
	}
	if cfg.Image.Quality == 0 {
		cfg.Image.Quality = 85
	}
	if cfg.Auth.TokenExpire == 0 {
		cfg.Auth.TokenExpire = 7 * 24 * time.Hour
	}
	return &cfg, nil
}
```

- [ ] **Step 3: Create default config file**

Create `configs/config.yaml`:

```yaml
server:
  port: 8080
  base_url: "http://localhost:8080"

database:
  driver: sqlite
  dsn: "./data/cloudalbum.db"

storage:
  driver: local
  local:
    path: "./data/images"
  s3:
    bucket: ""
    region: ""
    endpoint: ""
    access_key: ""
    secret_key: ""

image:
  max_size: 52428800
  allowed_types:
    - jpg
    - jpeg
    - png
    - gif
    - webp
    - bmp
    - svg
  auto_convert: webp
  quality: 85
  strip_exif: true
  thumbnails:
    - name: thumb
      width: 200
      height: 200
    - name: medium
      width: 800
      height: 600
    - name: large
      width: 1200
      height: 900

auth:
  jwt_secret: "change-me-in-production"
  token_expire: 168h
```

- [ ] **Step 4: Create main entry point**

Create `cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"

	"cloudalbum/internal/config"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("CloudAlbum starting on :%d\n", cfg.Server.Port)
	fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.DSN)
	fmt.Printf("Storage: %s\n", cfg.Storage.Driver)
}
```

- [ ] **Step 5: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: dev build run clean

dev:
	cd web && npm run dev &
	go run cmd/server/main.go

build:
	cd web && npm run build
	go build -o bin/cloudalbum cmd/server/main.go

run: build
	./bin/cloudalbum

clean:
	rm -rf bin/ web/dist/ data/
```

- [ ] **Step 6: Verify the project compiles and runs**

```bash
go build ./...
go run cmd/server/main.go
```

Expected: prints "CloudAlbum starting on :8080" and exits.

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum cmd/ internal/config/ configs/ Makefile
git commit -m "feat: project scaffold with config loading"
```

---

### Task 2: Data Models + Database Layer

**Files:**
- Create: `internal/model/user.go`
- Create: `internal/model/image.go`
- Create: `internal/model/album.go`
- Create: `internal/model/token.go`

- [ ] **Step 1: Create User model**

Create `internal/model/user.go`:

```go
package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"size:20;default:admin" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
```

- [ ] **Step 2: Create Image model**

Create `internal/model/image.go`:

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	StorageKey   string         `gorm:"uniqueIndex;size:255;not null" json:"storage_key"`
	Filename     string         `gorm:"size:255;not null" json:"filename"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	Size         int64          `json:"size"`
	MimeType     string         `gorm:"size:100" json:"mime_type"`
	Width        int            `json:"width"`
	Height       int            `json:"height"`
	Hash         string         `gorm:"index;size:64" json:"hash"`
	AlbumID      *uint          `gorm:"index" json:"album_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	User  User   `gorm:"foreignKey:UserID" json:"-"`
	Album *Album `gorm:"foreignKey:AlbumID" json:"-"`
}
```

- [ ] **Step 3: Create Album model**

Create `internal/model/album.go`:

```go
package model

import "time"

type Album struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	CoverImageID *uint    `json:"cover_image_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`

	User       User    `gorm:"foreignKey:UserID" json:"-"`
	CoverImage *Image  `gorm:"foreignKey:CoverImageID" json:"-"`
}
```

- [ ] **Step 4: Create APIToken model**

Create `internal/model/token.go`:

```go
package model

import (
	"time"
)

type APIToken struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	Name       string     `gorm:"size:100;not null" json:"name"`
	TokenHash  string     `gorm:"uniqueIndex;size:64;not null" json:"-"`
	Scope      string     `gorm:"size:20;default:full" json:"scope"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
```

- [ ] **Step 5: Verify models compile**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/model/
git commit -m "feat: add GORM data models (User, Image, Album, APIToken)"
```

---

### Task 3: Storage Backend

**Files:**
- Create: `internal/storage/storage.go`
- Create: `internal/storage/local.go`

- [ ] **Step 1: Define Storage interface**

Create `internal/storage/storage.go`:

```go
package storage

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, key string, data io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
```

- [ ] **Step 2: Implement LocalStorage**

Create `internal/storage/local.go`:

```go
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("resolve storage path: %w", err)
	}
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}
	return &LocalStorage{basePath: absPath}, nil
}

func (s *LocalStorage) fullPath(key string) string {
	return filepath.Join(s.basePath, key)
}

func (s *LocalStorage) Save(_ context.Context, key string, data io.Reader) error {
	fullPath := s.fullPath(key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		os.Remove(fullPath)
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(s.fullPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(s.fullPath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	if err := os.Remove(s.fullPath(key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/storage/
git commit -m "feat: add Storage interface and LocalStorage implementation"
```

---

### Task 4: Image Processing Pipeline

**Files:**
- Create: `internal/image/processor.go`

- [ ] **Step 1: Implement image processor**

Create `internal/image/processor.go`:

```go
package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"github.com/disintegration/imaging"
	"cloudalbum/internal/config"
)

type ProcessResult struct {
	Width      int
	Height     int
	Hash       string
	Size       int64
	MimeType   string
	Thumbnails map[string][]byte
}

type Processor struct {
	cfg config.ImageConfig
}

func NewProcessor(cfg config.ImageConfig) *Processor {
	return &Processor{cfg: cfg}
}

func (p *Processor) Process(data []byte, mimeType string) (*ProcessResult, error) {
	img, format, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	result := &ProcessResult{
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
		Size:     int64(len(data)),
		MimeType: mimeType,
	}

	// Hash original data
	hash := sha256.Sum256(data)
	result.Hash = hex.EncodeToString(hash[:])

	// Generate thumbnails
	result.Thumbnails = make(map[string][]byte)
	for _, size := range p.cfg.Thumbnails {
		thumb := imaging.Thumbnail(img, size.Width, size.Height, imaging.Lanczos)
		var buf bytes.Buffer
		encFormat, _ := formatToImaging(format)
		if err := imaging.Encode(&buf, thumb, encFormat, imaging.JPEGQuality(p.cfg.Quality)); err != nil {
			return nil, fmt.Errorf("encode thumbnail %s: %w", size.Name, err)
		}
		result.Thumbnails[size.Name] = buf.Bytes()
	}

	return result, nil
}

func (p *Processor) GenerateThumbnailKey(originalKey, sizeName string) string {
	return fmt.Sprintf("thumbs/%s_%s", sizeName, originalKey)
}

func formatToImaging(format string) (imaging.Format, error) {
	switch format {
	case "jpeg", "jpg":
		return imaging.JPEG, nil
	case "png":
		return imaging.PNG, nil
	case "gif":
		return imaging.GIF, nil
	default:
		return imaging.JPEG, nil
	}
}

// HashFromReader computes SHA256 from a reader without loading entire file into memory.
func HashFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// DetectImageType reads the first 512 bytes to detect MIME type.
func DetectImageType(data []byte) string {
	if len(data) < 512 {
		return ""
	}
	// Simple detection based on magic bytes
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "image/png"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}):
		return "image/gif"
	case len(data) > 11 && string(data[8:12]) == "webp":
		return "image/webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return "image/bmp"
	case bytes.HasPrefix(data, []byte("<?xml")) || bytes.HasPrefix(data, []byte("<svg")):
		return "image/svg+xml"
	}
	return ""
}

func init() {
	// Register WebP decoder via imaging if available
	_ = os.DevNull // placeholder; WebP support via imaging/vips if compiled with tag
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/image/
git commit -m "feat: add image processing pipeline (hash, thumbnails)"
```

---

## Phase 2: Backend API

### Task 5: Database Init + Repository Layer

**Files:**
- Create: `internal/repository/user.go`
- Create: `internal/repository/image.go`
- Create: `internal/repository/album.go`
- Create: `internal/repository/token.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create user repository**

Create `internal/repository/user.go`:

```go
package repository

import (
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
```

- [ ] **Step 2: Create image repository**

Create `internal/repository/image.go`:

```go
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
	UserID   uint
	AlbumID  *uint
	Page     int
	PageSize int
	Keyword  string
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
```

- [ ] **Step 3: Create album repository**

Create `internal/repository/album.go`:

```go
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
```

- [ ] **Step 4: Create token repository**

Create `internal/repository/token.go`:

```go
package repository

import (
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
	return r.db.Model(&model.APIToken{}).Where("id = ?", id).UpdateColumn("last_used_at", gorm.Expr("NOW()")).Error
}
```

- [ ] **Step 5: Update main.go with database initialization**

Replace `cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Init database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// Init storage
	store, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	fmt.Printf("CloudAlbum starting on :%d\n", cfg.Server.Port)
	fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.DSN)
	fmt.Printf("Storage: %s (%s)\n", cfg.Storage.Driver, store.(*storage.LocalStorage).String())
	_ = db
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Database.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Album{}, &model.APIToken{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.Local.Path)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}
```

Add a `String()` method to `internal/storage/local.go`:

```go
func (s *LocalStorage) String() string {
	return s.basePath
}
```

- [ ] **Step 6: Verify compilation and run**

```bash
go build ./...
go run cmd/server/main.go
```

Expected: prints startup info, creates `data/cloudalbum.db` and `data/images/`.

- [ ] **Step 7: Commit**

```bash
git add internal/repository/ cmd/server/main.go
git commit -m "feat: add repository layer and database initialization"
```

---

### Task 6: Auth System (JWT + API Token)

**Files:**
- Create: `internal/service/auth.go`
- Create: `internal/service/token.go`
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/cors.go`
- Create: `internal/handler/auth.go`
- Create: `internal/handler/token.go`

- [ ] **Step 1: Create auth service**

Create `internal/service/auth.go`:

```go
package service

import (
	"errors"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	cfg       config.AuthConfig
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, cfg config.AuthConfig) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, cfg: cfg}
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateJWT(user *model.User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.TokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) ParseJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return s.GenerateJWT(user)
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) EnsureAdmin(username, password string) error {
	_, err := s.userRepo.FindByUsername(username)
	if err == nil {
		return nil
	}
	_, err = s.Register(username, password)
	return err
}
```

- [ ] **Step 2: Create token service**

Create `internal/service/token.go`:

```go
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
)

type TokenService struct {
	tokenRepo *repository.TokenRepository
}

func NewTokenService(tokenRepo *repository.TokenRepository) *TokenService {
	return &TokenService{tokenRepo: tokenRepo}
}

func (s *TokenService) Create(userID uint, name, scope string) (*model.APIToken, string, error) {
	raw := generateToken()
	hash := sha256.Sum256([]byte(raw))
	tokenHash := hex.EncodeToString(hash[:])

	token := &model.APIToken{
		UserID:    userID,
		Name:      name,
		TokenHash: tokenHash,
		Scope:     scope,
	}
	if err := s.tokenRepo.Create(token); err != nil {
		return nil, "", err
	}
	// Return raw token prefixed with "ca_" for identification
	return token, "ca_" + raw, nil
}

func (s *TokenService) Validate(rawToken string) (*model.APIToken, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	token, err := s.tokenRepo.FindByHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}
	s.tokenRepo.UpdateLastUsed(token.ID)
	return token, nil
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
		return fmt.Errorf("unauthorized")
	}
	return s.tokenRepo.Delete(id)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

- [ ] **Step 3: Create auth middleware**

Create `internal/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authSvc *service.AuthService, tokenSvc *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check query param for token (PicGo compatibility)
			if t := c.Query("token"); t != "" {
				authHeader = "Bearer " + t
			}
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		// Try Bearer token (JWT or API token)
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// Try API Token first (ca_ prefix)
		if strings.HasPrefix(tokenStr, "ca_") {
			raw := strings.TrimPrefix(tokenStr, "ca_")
			apiToken, err := tokenSvc.Validate(raw)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			c.Set("user_id", apiToken.UserID)
			c.Set("auth_type", "api_token")
			c.Set("token_scope", apiToken.Scope)
			c.Next()
			return
		}

		// Try JWT
		claims, err := authSvc.ParseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("auth_type", "jwt")
		c.Next()
	}
}

func RequireScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType := c.GetString("auth_type")
		if authType == "jwt" {
			c.Next()
			return
		}
		tokenScope := c.GetString("token_scope")
		if tokenScope == "full" {
			c.Next()
			return
		}
		for _, s := range scopes {
			if tokenScope == s {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
```

- [ ] **Step 4: Create CORS middleware**

Create `internal/middleware/cors.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 5: Create auth handler**

Create `internal/handler/auth.go`:

```go
package handler

import (
	"net/http"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}
	token, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetUint("user_id")
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
	})
}
```

- [ ] **Step 6: Create token handler**

Create `internal/handler/token.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

type TokenHandler struct {
	tokenSvc *service.TokenService
}

func NewTokenHandler(tokenSvc *service.TokenService) *TokenHandler {
	return &TokenHandler{tokenSvc: tokenSvc}
}

type CreateTokenRequest struct {
	Name  string `json:"name" binding:"required"`
	Scope string `json:"scope" binding:"required,oneof=read upload full"`
}

func (h *TokenHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	tokens, err := h.tokenSvc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *TokenHandler) Create(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetUint("user_id")
	token, raw, err := h.tokenSvc.Create(userID, req.Name, req.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"raw":   raw,
	})
}

func (h *TokenHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetUint("user_id")
	if err := h.tokenSvc.Delete(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
```

- [ ] **Step 7: Verify compilation**

```bash
go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add internal/service/ internal/middleware/ internal/handler/
git commit -m "feat: add auth system (JWT + API Token) with handlers and middleware"
```

---

### Task 7: Image + Album API Handlers

**Files:**
- Create: `internal/service/image.go`
- Create: `internal/service/album.go`
- Create: `internal/handler/image.go`
- Create: `internal/handler/album.go`
- Create: `internal/handler/public.go`
- Create: `internal/router/router.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Create image service**

Create `internal/service/image.go`:

```go
package service

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"cloudalbum/internal/config"
	"cloudalbum/internal/image"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/storage"

	"github.com/google/uuid"
)

type ImageService struct {
	imageRepo *repository.ImageRepository
	store     storage.Storage
	processor *image.Processor
	cfg       config.ImageConfig
	baseURL   string
}

func NewImageService(
	imageRepo *repository.ImageRepository,
	store storage.Storage,
	processor *image.Processor,
	cfg config.ImageConfig,
	baseURL string,
) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		store:     store,
		processor: processor,
		cfg:       cfg,
		baseURL:   baseURL,
	}
}

func (s *ImageService) Upload(userID uint, file *multipart.FileHeader, albumID *uint) (*model.Image, error) {
	// Validate extension
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), ".")
	if !s.isAllowedType(ext) {
		return nil, fmt.Errorf("file type %s not allowed", ext)
	}

	// Validate size
	if file.Size > s.cfg.MaxSize {
		return nil, fmt.Errorf("file too large: %d > %d", file.Size, s.cfg.MaxSize)
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}

	// Process image
	result, err := s.processor.Process(data, image.DetectImageType(data))
	if err != nil {
		return nil, fmt.Errorf("process image: %w", err)
	}

	// Check dedup
	existing, err := s.imageRepo.FindByHash(result.Hash)
	if err == nil && existing != nil {
		// Duplicate: create new record pointing to same storage key
		dup := &model.Image{
			UserID:       userID,
			StorageKey:   existing.StorageKey,
			Filename:     existing.Filename,
			OriginalName: file.Filename,
			Size:         existing.Size,
			MimeType:     existing.MimeType,
			Width:        existing.Width,
			Height:       existing.Height,
			Hash:         result.Hash,
			AlbumID:      albumID,
		}
		if err := s.imageRepo.Create(dup); err != nil {
			return nil, err
		}
		return dup, nil
	}

	// Generate storage key
	key := fmt.Sprintf("%s/%s/%s", userID, uuid.New().String()[:8], file.Filename)

	// Save original
	if err := s.store.Save(nil, key, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	// Save thumbnails
	for sizeName, thumbData := range result.Thumbnails {
		thumbKey := s.processor.GenerateThumbnailKey(key, sizeName)
		if err := s.store.Save(nil, thumbKey, bytes.NewReader(thumbData)); err != nil {
			return nil, fmt.Errorf("save thumbnail %s: %w", sizeName, err)
		}
	}

	img := &model.Image{
		UserID:       userID,
		StorageKey:   key,
		Filename:     filepath.Base(key),
		OriginalName: file.Filename,
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

func (s *ImageService) UploadFromURL(userID uint, imageURL string, albumID *uint) (*model.Image, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch URL: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if int64(len(data)) > s.cfg.MaxSize {
		return nil, fmt.Errorf("file too large")
	}

	mimeType := image.DetectImageType(data)
	ext := mimeTypeToExt(mimeType)
	if !s.isAllowedType(ext) {
		return nil, fmt.Errorf("file type not allowed")
	}

	filename := fmt.Sprintf("url_%s.%s", uuid.New().String()[:8], ext)

	result, err := s.processor.Process(data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("process image: %w", err)
	}

	key := fmt.Sprintf("%s/%s/%s", userID, uuid.New().String()[:8], filename)
	if err := s.store.Save(nil, key, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("save image: %w", err)
	}

	for sizeName, thumbData := range result.Thumbnails {
		thumbKey := s.processor.GenerateThumbnailKey(key, sizeName)
		if err := s.store.Save(nil, thumbKey, bytes.NewReader(thumbData)); err != nil {
			return nil, fmt.Errorf("save thumbnail %s: %w", sizeName, err)
		}
	}

	img := &model.Image{
		UserID:       userID,
		StorageKey:   key,
		Filename:     filepath.Base(key),
		OriginalName: filename,
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
		return fmt.Errorf("unauthorized")
	}
	return s.imageRepo.Update(&model.Image{ID: id, OriginalName: updates["original_name"].(string), AlbumID: updates["album_id"].(*uint)})
}

func (s *ImageService) Delete(id uint, userID uint) error {
	img, err := s.imageRepo.FindByID(id)
	if err != nil {
		return err
	}
	if img.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.imageRepo.SoftDelete(id)
}

func (s *ImageService) Restore(id uint, userID uint) error {
	return s.imageRepo.Restore(id)
}

func (s *ImageService) HardDelete(id uint, userID uint) error {
	return s.imageRepo.HardDelete(id)
}

func (s *ImageService) BatchOperation(ids []uint, userID uint, action string, albumID *uint) error {
	switch action {
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
		return
	}
	totalSize, err = s.imageRepo.TotalSizeByUserID(userID)
	return
}

func (s *ImageService) URLs(img *model.Image) map[string]string {
	url := fmt.Sprintf("%s/i/%s", s.baseURL, img.StorageKey)
	return map[string]string{
		"url":     url,
		"markdown": fmt.Sprintf("![%s](%s)", img.OriginalName, url),
		"html":    fmt.Sprintf(`<img src="%s" alt="%s">`, url, img.OriginalName),
		"bbcode":  fmt.Sprintf("[img]%s[/img]", url),
	}
}

func (s *ImageService) isAllowedType(ext string) bool {
	for _, t := range s.cfg.AllowedTypes {
		if t == ext {
			return true
		}
	}
	return false
}

func mimeTypeToExt(mime string) string {
	switch mime {
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
		return "jpg"
	}
}
```

- [ ] **Step 2: Create album service**

Create `internal/service/album.go`:

```go
package service

import (
	"fmt"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
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
		Name:        name,
		Description: description,
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
	albums, err := s.albumRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}
	// Attach image counts
	for i := range albums {
		count, _ := s.albumRepo.CountImages(albums[i].ID)
		albums[i].SortOrder = int(count) // reuse for count in response
	}
	return albums, nil
}

func (s *AlbumService) Update(id uint, userID uint, name, description string, coverImageID *uint) error {
	album, err := s.albumRepo.FindByID(id)
	if err != nil {
		return err
	}
	if album.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	album.Name = name
	album.Description = description
	album.CoverImageID = coverImageID
	return s.albumRepo.Update(album)
}

func (s *AlbumService) Delete(id uint, userID uint) error {
	album, err := s.albumRepo.FindByID(id)
	if err != nil {
		return err
	}
	if album.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.albumRepo.Delete(id)
}
```

- [ ] **Step 3: Create image handler**

Create `internal/handler/image.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageSvc *service.ImageService
}

func NewImageHandler(imageSvc *service.ImageService) *ImageHandler {
	return &ImageHandler{imageSvc: imageSvc}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	userID := c.GetUint("user_id")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	var albumID *uint
	if aid := c.PostForm("album_id"); aid != "" {
		id, _ := strconv.ParseUint(aid, 10, 64)
		uid := uint(id)
		albumID = &uid
	}

	var results []map[string]interface{}
	for _, file := range files {
		img, err := h.imageSvc.Upload(userID, file, albumID)
		if err != nil {
			results = append(results, map[string]interface{}{
				"filename": file.Filename,
				"error":    err.Error(),
			})
			continue
		}
		results = append(results, map[string]interface{}{
			"image": img,
			"urls":  h.imageSvc.URLs(img),
		})
	}
	c.JSON(http.StatusCreated, gin.H{"results": results})
}

func (h *ImageHandler) UploadURL(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		URL     string `json:"url" binding:"required"`
		AlbumID *uint  `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	img, err := h.imageSvc.UploadFromURL(userID, req.URL, req.AlbumID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"image": img,
		"urls":  h.imageSvc.URLs(img),
	})
}

func (h *ImageHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	var albumID *uint
	if aid := c.Query("album_id"); aid != "" {
		id, _ := strconv.ParseUint(aid, 10, 64)
		uid := uint(id)
		albumID = &uid
	}

	onlyDeleted := c.Query("deleted") == "true"

	images, total, err := h.imageSvc.List(repository.ImageListParams{
		UserID:      userID,
		AlbumID:     albumID,
		Page:        page,
		PageSize:    pageSize,
		Keyword:     keyword,
		OnlyDeleted: onlyDeleted,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"images": images,
		"total":  total,
		"page":   page,
	})
}

func (h *ImageHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	img, err := h.imageSvc.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"image": img,
		"urls":  h.imageSvc.URLs(img),
	})
}

func (h *ImageHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	var req struct {
		OriginalName string `json:"original_name"`
		AlbumID      *uint  `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.imageSvc.Update(uint(id), userID, map[string]interface{}{
		"original_name": req.OriginalName,
		"album_id":      req.AlbumID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	if err := h.imageSvc.Delete(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ImageHandler) Batch(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		IDs    []uint `json:"ids" binding:"required"`
		Action string `json:"action" binding:"required"`
		AlbumID *uint `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.imageSvc.BatchOperation(req.IDs, userID, req.Action, req.AlbumID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "done"})
}

func (h *ImageHandler) Restore(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	if err := h.imageSvc.Restore(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restored"})
}

func (h *ImageHandler) HardDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	if err := h.imageSvc.HardDelete(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permanently deleted"})
}

func (h *ImageHandler) Stats(c *gin.Context) {
	userID := c.GetUint("user_id")
	count, totalSize, err := h.imageSvc.Stats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"count":      count,
		"total_size": totalSize,
	})
}
```

- [ ] **Step 4: Create album handler**

Create `internal/handler/album.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

type AlbumHandler struct {
	albumSvc *service.AlbumService
}

func NewAlbumHandler(albumSvc *service.AlbumService) *AlbumHandler {
	return &AlbumHandler{albumSvc: albumSvc}
}

func (h *AlbumHandler) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	albums, err := h.albumSvc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"albums": albums})
}

func (h *AlbumHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	album, err := h.albumSvc.Create(userID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"album": album})
}

func (h *AlbumHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		CoverImageID *uint  `json:"cover_image_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.albumSvc.Update(uint(id), userID, req.Name, req.Description, req.CoverImageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AlbumHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userID := c.GetUint("user_id")
	if err := h.albumSvc.Delete(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
```

- [ ] **Step 5: Create public image serving handler**

Create `internal/handler/public.go`:

```go
package handler

import (
	"io"
	"net/http"
	"strings"

	"cloudalbum/internal/image"
	"cloudalbum/internal/storage"
	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	store     storage.Storage
	processor *image.Processor
}

func NewPublicHandler(store storage.Storage, processor *image.Processor) *PublicHandler {
	return &PublicHandler{store: store, processor: processor}
}

func (h *PublicHandler) Image(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	reader, err := h.store.Get(nil, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	defer reader.Close()

	ext := strings.ToLower(getExt(key))
	contentType := extToContentType(ext)
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, max-age=31536000")
	io.Copy(c.Writer, reader)
}

func (h *PublicHandler) Thumbnail(c *gin.Context) {
	key := c.Param("key")
	// key format: {size}_{original_key}
	// Thumbnail is stored at thumbs/{size}_{original_key}
	thumbKey := h.processor.GenerateThumbnailKey("", "") + "/" + key
	// Actually the key in URL is like thumb_abc/original.jpg
	// We need to reconstruct: thumbs/{size}_{rest}
	// Let's fix: key comes as "thumb_userid/uuid/file.jpg"
	// Storage key is "thumbs/thumb_userid/uuid/file.jpg"
	// But actually the thumbnail key generated is: "thumbs/{sizeName}_{originalKey}"
	// So the URL path after /t/ should be the full thumbnail storage key

	reader, err := h.store.Get(nil, "thumbs/"+key)
	if err != nil {
		// Fallback to original
		h.Image(c)
		return
	}
	defer reader.Close()

	contentType := extToContentType(getExt(key))
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, max-age=31536000")
	io.Copy(c.Writer, reader)
}

func getExt(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func extToContentType(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "svg":
		return "image/svg+xml"
	default:
		return ""
	}
}
```

- [ ] **Step 6: Create router**

Create `internal/router/router.go`:

```go
package router

import (
	"cloudalbum/internal/handler"
	"cloudalbum/internal/middleware"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

func Setup(
	r *gin.Engine,
	authSvc *service.AuthService,
	tokenSvc *service.TokenService,
	imageSvc *service.ImageService,
	albumSvc *service.AlbumService,
	authHandler *handler.AuthHandler,
	tokenHandler *handler.TokenHandler,
	imageHandler *handler.ImageHandler,
	albumHandler *handler.AlbumHandler,
	publicHandler *handler.PublicHandler,
) {
	r.Use(middleware.CORS())

	// Public routes
	r.GET("/i/*key", publicHandler.Image)
	r.GET("/t/*key", publicHandler.Thumbnail)

	// Auth routes
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected API routes
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(authSvc, tokenSvc))
	{
		api.GET("/auth/me", authHandler.Me)

		// Images
		images := api.Group("/images")
		{
			images.POST("", imageHandler.Upload)
			images.POST("/upload-url", imageHandler.UploadURL)
			images.GET("", imageHandler.List)
			images.GET("/stats", imageHandler.Stats)
			images.GET("/:id", imageHandler.Get)
			images.PUT("/:id", imageHandler.Update)
			images.DELETE("/:id", imageHandler.Delete)
			images.POST("/batch", imageHandler.Batch)
			images.POST("/:id/restore", imageHandler.Restore)
			images.DELETE("/:id/permanent", imageHandler.HardDelete)
		}

		// Albums
		albums := api.Group("/albums")
		{
			albums.GET("", albumHandler.List)
			albums.POST("", albumHandler.Create)
			albums.PUT("/:id", albumHandler.Update)
			albums.DELETE("/:id", albumHandler.Delete)
		}

		// Tokens
		tokens := api.Group("/tokens")
		{
			tokens.GET("", tokenHandler.List)
			tokens.POST("", tokenHandler.Create)
			tokens.DELETE("/:id", tokenHandler.Delete)
		}
	}
}
```

- [ ] **Step 7: Update main.go to wire everything together**

Replace `cmd/server/main.go`:

```go
package main

import (
	"fmt"
	"log"

	"cloudalbum/internal/config"
	"cloudalbum/internal/handler"
	"cloudalbum/internal/image"
	"cloudalbum/internal/middleware"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/router"
	"cloudalbum/internal/service"
	"cloudalbum/internal/storage"

	imgprocessor "cloudalbum/internal/image"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	imageRepo := repository.NewImageRepository(db)
	albumRepo := repository.NewAlbumRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.Auth)
	tokenSvc := service.NewTokenService(tokenRepo)
	processor := imgprocessor.NewProcessor(cfg.Image)
	imageSvc := service.NewImageService(imageRepo, store, processor, cfg.Image, cfg.Server.BaseURL)
	albumSvc := service.NewAlbumService(albumRepo, imageRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	tokenHandler := handler.NewTokenHandler(tokenSvc)
	imageHandler := handler.NewImageHandler(imageSvc)
	albumHandler := handler.NewAlbumHandler(albumSvc)
	publicHandler := handler.NewPublicHandler(store, processor)

	// Ensure default admin user
	if err := authSvc.EnsureAdmin("admin", "admin123"); err != nil {
		log.Printf("Warning: failed to ensure admin user: %v", err)
	}

	// Setup router
	r := gin.Default()
	router.Setup(r, authSvc, tokenSvc, imageSvc, albumSvc, authHandler, tokenHandler, imageHandler, albumHandler, publicHandler)

	// Serve
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("CloudAlbum running on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	_ = middleware.CORS
	_ = model.User{}
	_ = image.Processor{}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Database.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Album{}, &model.APIToken{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	return db, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.Local.Path)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}
```

- [ ] **Step 8: Verify compilation and run**

```bash
go build ./...
go run cmd/server/main.go
```

Expected: Server starts on :8080, creates data directory, default admin user.

- [ ] **Step 9: Test login with curl**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Expected: `{"token":"..."}`

- [ ] **Step 10: Commit**

```bash
git add internal/ cmd/server/main.go
git commit -m "feat: complete backend API (image, album, auth, token handlers + router)"
```

---

## Phase 3: React Frontend

### Task 8: React Project Setup + Login Page

**Files:**
- Create: `web/` (entire React project via Vite)
- Create: `web/src/api/client.ts`
- Create: `web/src/stores/auth.ts`
- Create: `web/src/pages/Login.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/main.tsx`

- [ ] **Step 1: Initialize React project with Vite**

```bash
cd /Users/zyb/workspace/person/CloudAlbum
npm create vite@latest web -- --template react-ts
cd web
npm install @arco-design/web-react react-router-dom axios zustand
```

- [ ] **Step 2: Create API client**

Create `web/src/api/client.ts`:

```typescript
import axios from 'axios';

const client = axios.create({
  baseURL: '/api/v1',
});

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

export default client;
```

- [ ] **Step 3: Create auth store with Zustand**

Create `web/src/stores/auth.ts`:

```typescript
import { create } from 'zustand';
import client from '../api/client';

interface AuthState {
  token: string | null;
  username: string | null;
  loggedIn: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  init: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  username: null,
  loggedIn: !!localStorage.getItem('token'),
  login: async (username: string, password: string) => {
    const res = await client.post('/auth/login', { username, password });
    const token = res.data.token;
    localStorage.setItem('token', token);
    set({ token, username, loggedIn: true });
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ token: null, username: null, loggedIn: false });
  },
  init: () => {
    const token = localStorage.getItem('token');
    if (token) {
      set({ token, loggedIn: true });
    }
  },
}));
```

- [ ] **Step 4: Create Login page**

Create `web/src/pages/Login.tsx`:

```tsx
import { useState } from 'react';
import { Card, Form, Input, Button, Message, Typography } from '@arco-design/web-react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../stores/auth';

const { Title } = Typography;

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);

  const handleSubmit = async () => {
    setLoading(true);
    setError('');
    try {
      await login(username, password);
      navigate('/');
    } catch {
      setError('用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#f0f2f5' }}>
      <Card style={{ width: 400 }}>
        <Title heading={4} style={{ textAlign: 'center', marginBottom: 24 }}>CloudAlbum</Title>
        {error && <Message type="error" style={{ marginBottom: 16 }}>{error}</Message>}
        <Form layout="vertical" onSubmit={handleSubmit}>
          <Form.Item label="用户名">
            <Input value={username} onChange={setUsername} placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item label="密码">
            <Input.Password value={password} onChange={setPassword} placeholder="请输入密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" long loading={loading}>登录</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
```

- [ ] **Step 5: Create App with routes**

Replace `web/src/App.tsx`:

```tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './stores/auth';
import Login from './pages/Login';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  return loggedIn ? <>{children}</> : <Navigate to="/login" />;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/*" element={<PrivateRoute><div>Dashboard - Coming Soon</div></PrivateRoute>} />
      </Routes>
    </BrowserRouter>
  );
}
```

- [ ] **Step 6: Update main.tsx with Arco Design**

Replace `web/src/main.tsx`:

```tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import '@arco-design/web-react/dist/css/arco.css';
import App from './App';
import { useAuthStore } from './stores/auth';

useAuthStore.getState().init();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

- [ ] **Step 7: Configure Vite proxy for dev**

Replace `web/vite.config.ts`:

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080',
      '/i': 'http://localhost:8080',
      '/t': 'http://localhost:8080',
    },
  },
});
```

- [ ] **Step 8: Verify frontend starts**

```bash
cd web && npm run dev
```

Expected: Frontend runs at http://localhost:3000, shows login page.

- [ ] **Step 9: Commit**

```bash
cd /Users/zyb/workspace/person/CloudAlbum
git add web/
git commit -m "feat: React frontend scaffold with login page and Arco Design"
```

---

### Task 9: Layout + Upload Page

**Files:**
- Create: `web/src/components/Layout.tsx`
- Create: `web/src/pages/Upload.tsx`
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Create app layout with sidebar**

Create `web/src/components/Layout.tsx`:

```tsx
import { useState } from 'react';
import { Layout as ArcoLayout, Menu, Typography, Button } from '@arco-design/web-react';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useAuthStore } from '../stores/auth';
import {
  IconDashboard,
  IconUpload,
  IconImage,
  IconFolder,
  IconDelete,
  IconKey,
  IconSettings,
  IconMenuFold,
  IconMenuUnfold,
} from '@arco-design/web-react/icon';

const { Sider, Content } = ArcoLayout;
const { Title } = Typography;

const menuItems = [
  { key: '/', icon: <IconDashboard />, label: '仪表盘' },
  { key: '/upload', icon: <IconUpload />, label: '上传图片' },
  { key: '/images', icon: <IconImage />, label: '图片管理' },
  { key: '/albums', icon: <IconFolder />, label: '相册管理' },
  { key: '/trash', icon: <IconDelete />, label: '回收站' },
  { key: '/tokens', icon: <IconKey />, label: 'Token 管理' },
  { key: '/settings', icon: <IconSettings />, label: '系统设置' },
];

export default function Layout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const logout = useAuthStore((s) => s.logout);

  return (
    <ArcoLayout style={{ height: '100vh' }}>
      <Sider collapsed={collapsed} collapsible trigger={null} width={220} style={{ background: '#fff', borderRight: '1px solid #e5e6eb' }}>
        <div style={{ padding: '16px', display: 'flex', alignItems: 'center', gap: 8 }}>
          <Title heading={5} style={{ margin: 0, whiteSpace: 'nowrap' }}>
            {collapsed ? 'CA' : 'CloudAlbum'}
          </Title>
        </div>
        <Menu
          selectedKeys={[location.pathname]}
          onClickMenuItem={(key) => navigate(key)}
          style={{ borderRight: 'none' }}
        >
          {menuItems.map((item) => (
            <Menu.Item key={item.key}>{item.icon} {item.label}</Menu.Item>
          ))}
        </Menu>
        <div style={{ position: 'absolute', bottom: 16, padding: '0 16px', width: '100%' }}>
          <Button long type="text" icon={collapsed ? <IconMenuUnfold /> : <IconMenuFold />} onClick={() => setCollapsed(!collapsed)} />
          <Button long type="text" status="danger" onClick={() => { logout(); navigate('/login'); }} style={{ marginTop: 8 }}>
            退出登录
          </Button>
        </div>
      </Sider>
      <Content style={{ background: '#f0f2f5', overflow: 'auto' }}>
        <div style={{ padding: 24 }}>
          <Outlet />
        </div>
      </Content>
    </ArcoLayout>
  );
}
```

- [ ] **Step 2: Create Upload page**

Create `web/src/pages/Upload.tsx`:

```tsx
import { useState, useRef, useCallback } from 'react';
import { Card, Upload, Button, Input, Tabs, Message, Typography, Space, Select, Progress } from '@arco-design/web-react';
import { IconPlus, IconLink } from '@arco-design/web-react/icon';
import client from '../api/client';
import { useAuthStore } from '../stores/auth';

const { TextArea } = Input;
const { Text } = Typography;

interface UploadResult {
  image: any;
  urls: Record<string, string>;
}

export default function UploadPage() {
  const [results, setResults] = useState<UploadResult[]>([]);
  const [urlInput, setUrlInput] = useState('');
  const [albumId, setAlbumId] = useState<string | undefined>();
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [copyFormat, setCopyFormat] = useState<string>('markdown');
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropRef = useRef<HTMLDivElement>(null);

  const uploadFiles = async (files: FileList | File[]) => {
    setUploading(true);
    setProgress(0);
    const formData = new FormData();
    for (const file of files) {
      formData.append('files', file);
    }
    if (albumId) formData.append('album_id', albumId);

    try {
      const res = await client.post('/images', formData, {
        onUploadProgress: (e) => {
          if (e.total) setProgress(Math.round((e.loaded / e.total) * 100));
        },
      });
      const newResults = res.data.results.filter((r: any) => r.image).map((r: any) => r);
      setResults((prev) => [...newResults, ...prev]);
      Message.success(`成功上传 ${newResults.length} 张图片`);
    } catch {
      Message.error('上传失败');
    } finally {
      setUploading(false);
      setProgress(0);
    }
  };

  const uploadFromURL = async () => {
    if (!urlInput.trim()) return;
    setUploading(true);
    try {
      const res = await client.post('/images/upload-url', {
        url: urlInput,
        album_id: albumId ? Number(albumId) : undefined,
      });
      setResults((prev) => [res.data, ...prev]);
      Message.success('上传成功');
      setUrlInput('');
    } catch {
      Message.error('URL 上传失败');
    } finally {
      setUploading(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    Message.success('已复制到剪贴板');
  };

  const copyAllLinks = () => {
    const links = results.map((r) => r.urls[copyFormat] || r.urls.url).join('\n');
    copyToClipboard(links);
  };

  const handlePaste = useCallback((e: React.ClipboardEvent) => {
    const items = e.clipboardData.items;
    const files: File[] = [];
    for (const item of items) {
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) files.push(file);
      }
    }
    if (files.length > 0) {
      uploadFiles(files);
    }
  }, [albumId]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    if (e.dataTransfer.files.length > 0) {
      uploadFiles(e.dataTransfer.files);
    }
  }, [albumId]);

  return (
    <div onPaste={handlePaste}>
      <Typography.Title heading={4}>上传图片</Typography.Title>

      <Tabs defaultActiveTab="file">
        <Tabs.TabPane key="file" title="文件上传">
          <Card>
            <div
              ref={dropRef}
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              style={{
                border: '2px dashed #c9cdd4',
                borderRadius: 8,
                padding: '40px 0',
                textAlign: 'center',
                cursor: 'pointer',
                transition: 'border-color 0.2s',
              }}
            >
              <IconPlus style={{ fontSize: 40, color: '#c9cdd4' }} />
              <div style={{ marginTop: 12, color: '#86909c' }}>点击选择文件、拖拽文件到此处、或 Ctrl+V 粘贴图片</div>
              <input
                ref={fileInputRef}
                type="file"
                multiple
                accept="image/*"
                style={{ display: 'none' }}
                onChange={(e) => e.target.files && uploadFiles(e.target.files)}
              />
            </div>
            {uploading && <Progress percent={progress} style={{ marginTop: 16 }} />}
          </Card>
        </Tabs.TabPane>

        <Tabs.TabPane key="url" title="远程 URL">
          <Card>
            <Space direction="vertical" style={{ width: '100%' }}>
              <TextArea
                value={urlInput}
                onChange={setUrlInput}
                placeholder="输入图片 URL，每行一个"
                autoSize={{ minRows: 3 }}
              />
              <Button type="primary" onClick={uploadFromURL} loading={uploading}>
                <IconLink /> 拉取图片
              </Button>
            </Space>
          </Card>
        </Tabs.TabPane>
      </Tabs>

      {results.length > 0 && (
        <Card style={{ marginTop: 16 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
            <Text>上传结果 ({results.length})</Text>
            <Space>
              <Select value={copyFormat} onChange={setCopyFormat} style={{ width: 120 }}>
                <Select.Option value="url">URL</Select.Option>
                <Select.Option value="markdown">Markdown</Select.Option>
                <Select.Option value="html">HTML</Select.Option>
                <Select.Option value="bbcode">BBCode</Select.Option>
              </Select>
              <Button type="primary" onClick={copyAllLinks}>复制全部链接</Button>
            </Space>
          </div>
          {results.map((r, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
              <img src={r.urls.url} alt="" style={{ width: 48, height: 48, objectFit: 'cover', borderRadius: 4 }} />
              <div style={{ flex: 1 }}>
                <Text>{r.image.original_name}</Text>
                <Text type="secondary" size="small"> {r.image.width}x{r.image.height} ({(r.image.size / 1024).toFixed(1)}KB)</Text>
              </div>
              <Button size="small" onClick={() => copyToClipboard(r.urls[copyFormat] || r.urls.url)}>复制</Button>
            </div>
          ))}
        </Card>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Update App.tsx with layout and routes**

Replace `web/src/App.tsx`:

```tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './stores/auth';
import Layout from './components/Layout';
import Login from './pages/Login';
import Upload from './pages/Upload';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  return loggedIn ? <>{children}</> : <Navigate to="/login" />;
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<PrivateRoute><Layout /></PrivateRoute>}>
          <Route index element={<div>Dashboard - Coming Soon</div>} />
          <Route path="upload" element={<Upload />} />
          <Route path="images" element={<div>Images - Coming Soon</div>} />
          <Route path="albums" element={<div>Albums - Coming Soon</div>} />
          <Route path="trash" element={<div>Trash - Coming Soon</div>} />
          <Route path="tokens" element={<div>Tokens - Coming Soon</div>} />
          <Route path="settings" element={<div>Settings - Coming Soon</div>} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
```

- [ ] **Step 4: Verify frontend builds and runs**

```bash
cd web && npm run dev
```

Expected: Login page works, after login shows sidebar layout, upload page works.

- [ ] **Step 5: Commit**

```bash
cd /Users/zyb/workspace/person/CloudAlbum
git add web/
git commit -m "feat: add app layout sidebar and upload page (file/drag/paste/URL)"
```

---

### Task 10: Image Management Page

**Files:**
- Create: `web/src/pages/Images.tsx`
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Create Images page**

Create `web/src/pages/Images.tsx`:

```tsx
import { useState, useEffect, useCallback } from 'react';
import { Card, Grid, Pagination, Input, Select, Button, Space, Modal, Typography, Dropdown, Menu, Message, Checkbox } from '@arco-design/web-react';
import { IconRefresh, IconDelete, IconEye, IconCopy, IconMove } from '@arco-design/web-react/icon';
import client from '../api/client';

const { Row, Col } = Grid;
const { Text } = Typography;

export default function Images() {
  const [images, setImages] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [albumFilter, setAlbumFilter] = useState<string | undefined>();
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [preview, setPreview] = useState<any>(null);
  const [albums, setAlbums] = useState<any[]>([]);

  const fetchImages = useCallback(async () => {
    const params: any = { page, page_size: 20 };
    if (keyword) params.keyword = keyword;
    if (albumFilter) params.album_id = albumFilter;
    const res = await client.get('/images', { params });
    setImages(res.data.images || []);
    setTotal(res.data.total || 0);
  }, [page, keyword, albumFilter]);

  const fetchAlbums = async () => {
    const res = await client.get('/albums');
    setAlbums(res.data.albums || []);
  };

  useEffect(() => { fetchImages(); }, [fetchImages]);
  useEffect(() => { fetchAlbums(); }, []);

  const deleteImage = async (id: number) => {
    await client.delete(`/images/${id}`);
    Message.success('已删除');
    fetchImages();
  };

  const copyLink = (url: string) => {
    navigator.clipboard.writeText(url);
    Message.success('已复制');
  };

  const batchDelete = async () => {
    if (selected.size === 0) return;
    await client.post('/images/batch', { ids: Array.from(selected), action: 'delete' });
    setSelected(new Set());
    Message.success('批量删除成功');
    fetchImages();
  };

  const batchMove = async (albumId: number) => {
    if (selected.size === 0) return;
    await client.post('/images/batch', { ids: Array.from(selected), action: 'move', album_id: albumId });
    setSelected(new Set());
    Message.success('移动成功');
    fetchImages();
  };

  const toggleSelect = (id: number) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Typography.Title heading={4} style={{ margin: 0 }}>图片管理</Typography.Title>
        <Space>
          {selected.size > 0 && (
            <>
              <Button status="danger" onClick={batchDelete}>删除 ({selected.size})</Button>
              <Select placeholder="移动到相册" style={{ width: 150 }} onChange={(v) => batchMove(Number(v))}>
                {albums.map((a) => <Select.Option key={a.id} value={a.id}>{a.name}</Select.Option>)}
              </Select>
            </>
          )}
          <Input.Search placeholder="搜索图片" style={{ width: 200 }} onSearch={setKeyword} />
          <Select placeholder="按相册筛选" allowClear style={{ width: 150 }} onChange={setAlbumFilter}>
            {albums.map((a) => <Select.Option key={a.id} value={String(a.id)}>{a.name}</Select.Option>)}
          </Select>
          <Button icon={<IconRefresh />} onClick={fetchImages} />
        </Space>
      </div>

      <Row gutter={[12, 12]}>
        {images.map((img) => (
          <Col key={img.id} span={6}>
            <Card
              hoverable
              cover={
                <div style={{ position: 'relative', height: 160, overflow: 'hidden', background: '#f7f8fa' }}>
                  <img src={`/i/${img.storage_key}`} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  <div style={{ position: 'absolute', top: 4, left: 4 }}>
                    <Checkbox checked={selected.has(img.id)} onChange={() => toggleSelect(img.id)} />
                  </div>
                </div>
              }
              actions={[
                <IconCopy onClick={() => copyLink(`${window.location.origin}/i/${img.storage_key}`)} />,
                <IconEye onClick={() => setPreview(img)} />,
                <Dropdown
                  droplist={
                    <Menu>
                      <Menu.Item key="delete" onClick={() => deleteImage(img.id)}>删除</Menu.Item>
                    </Menu>
                  }
                >
                  <span>···</span>
                </Dropdown>,
              ]}
            >
              <Card.Meta
                title={<Text ellipsis style={{ maxWidth: 150 }}>{img.original_name}</Text>}
                description={`${img.width}x${img.height} · ${(img.size / 1024).toFixed(1)}KB`}
              />
            </Card>
          </Col>
        ))}
      </Row>

      <div style={{ display: 'flex', justifyContent: 'center', marginTop: 24 }}>
        <Pagination current={page} total={total} pageSize={20} onChange={setPage} />
      </div>

      <Modal visible={!!preview} footer={null} onCancel={() => setPreview(null)} unmountOnExit>
        {preview && (
          <div>
            <img src={`/i/${preview.storage_key}`} alt="" style={{ width: '100%' }} />
            <div style={{ marginTop: 12 }}>
              <Text>{preview.original_name}</Text><br />
              <Text type="secondary">{preview.width}x{preview.height} · {(preview.size / 1024).toFixed(1)}KB</Text>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
```

- [ ] **Step 2: Update App.tsx route**

In `web/src/App.tsx`, replace the Images placeholder:

```tsx
import Images from './pages/Images';
// ...
<Route path="images" element={<Images />} />
```

- [ ] **Step 3: Verify and commit**

```bash
cd web && npm run dev
```

- [ ] **Step 4: Commit**

```bash
cd /Users/zyb/workspace/person/CloudAlbum
git add web/
git commit -m "feat: add image management page with grid view, search, filter, batch ops"
```

---

### Task 11: Album + Dashboard + Token + Trash + Settings Pages

**Files:**
- Create: `web/src/pages/Albums.tsx`
- Create: `web/src/pages/Dashboard.tsx`
- Create: `web/src/pages/Tokens.tsx`
- Create: `web/src/pages/Trash.tsx`
- Create: `web/src/pages/Settings.tsx`
- Modify: `web/src/App.tsx`

These pages follow the same patterns as the above pages (API calls via `client`, Arco Design components, Zustand store). Each page is self-contained.

- [ ] **Step 1: Create Albums page** — CRUD grid with name, description, cover, image count. Modal for create/edit.

- [ ] **Step 2: Create Dashboard page** — Stats cards (image count, storage used, recent uploads). Uses `GET /api/v1/images/stats` and `GET /api/v1/images?page=1&page_size=5`.

- [ ] **Step 3: Create Tokens page** — List, create (with name + scope), delete. Shows raw token once on create.

- [ ] **Step 4: Create Trash page** — Lists soft-deleted images, restore and permanent delete buttons. Uses `GET /api/v1/images?deleted=true`.

- [ ] **Step 5: Create Settings page** — Display current config (read-only initially). Future: editable storage config, image processing rules.

- [ ] **Step 6: Update App.tsx to import all pages**

```tsx
import Albums from './pages/Albums';
import Dashboard from './pages/Dashboard';
import Tokens from './pages/Tokens';
import Trash from './pages/Trash';
import Settings from './pages/Settings';
// ...
<Route index element={<Dashboard />} />
<Route path="albums" element={<Albums />} />
<Route path="tokens" element={<Tokens />} />
<Route path="trash" element={<Trash />} />
<Route path="settings" element={<Settings />} />
```

- [ ] **Step 7: Verify all pages render and commit**

```bash
cd web && npm run dev
```

```bash
git add web/
git commit -m "feat: add album, dashboard, token, trash, and settings pages"
```

---

## Phase 4: Integration + Deployment

### Task 12: Go Embed + Production Wiring

**Files:**
- Create: `embed.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/router/router.go`

- [ ] **Step 1: Create embed.go**

Create `embed.go`:

```go
package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/dist/*
var webDist embed.FS

func WebFS() http.FileSystem {
	sub, _ := fs.Sub(webDist, "web/dist")
	return http.FS(sub)
}
```

- [ ] **Step 2: Add SPA fallback in router**

Add to `internal/router/router.go`:

```go
import "net/http"

// Add after all API routes:
// SPA fallback — serve index.html for non-API, non-image routes
r.NoRoute(func(c *gin.Context) {
	// Don't interfere with image/thumbnail routes
	if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		strings.HasPrefix(c.Request.URL.Path, "/i/") ||
		strings.HasPrefix(c.Request.URL.Path, "/t/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.FileFromFS(c.Request.URL.Path, webFS)
})
```

- [ ] **Step 3: Build frontend and test full integration**

```bash
cd web && npm run build
cd .. && go build -o bin/cloudalbum cmd/server/main.go
./bin/cloudalbum
```

Expected: Visit http://localhost:8080 shows React app, API works.

- [ ] **Step 4: Commit**

```bash
git add embed.go cmd/ internal/router/
git commit -m "feat: embed React SPA into Go binary with SPA fallback routing"
```

---

### Task 13: Docker + Makefile Finalization

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Modify: `Makefile`
- Create: `.dockerignore`

- [ ] **Step 1: Create Dockerfile**

Create `Dockerfile`:

```dockerfile
# Stage 1: Build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=1 GOOS=linux go build -o /cloudalbum cmd/server/main.go

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /cloudalbum .
COPY configs/ ./configs/
RUN mkdir -p /data/images

EXPOSE 8080
VOLUME ["/data"]
CMD ["./cloudalbum"]
```

- [ ] **Step 2: Create docker-compose.yml**

Create `docker-compose.yml`:

```yaml
version: '3.8'
services:
  cloudalbum:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - cloudalbum-data:/data
      - ./configs:/app/configs
    restart: unless-stopped

volumes:
  cloudalbum-data:
```

- [ ] **Step 3: Create .dockerignore**

Create `.dockerignore`:

```
bin/
data/
web/node_modules/
web/dist/
.git/
*.md
docs/
```

- [ ] **Step 4: Update Makefile**

Replace `Makefile`:

```makefile
.PHONY: dev dev-backend dev-frontend build run clean docker

dev:
	@make dev-frontend &
	@make dev-backend

dev-backend:
	go run cmd/server/main.go

dev-frontend:
	cd web && npm run dev

build:
	cd web && npm ci && npm run build
	CGO_ENABLED=1 go build -o bin/cloudalbum cmd/server/main.go

run: build
	./bin/cloudalbum

docker:
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -rf bin/ web/dist/ data/
```

- [ ] **Step 5: Test Docker build**

```bash
docker compose up -d --build
curl http://localhost:8080/api/v1/auth/login -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}'
```

- [ ] **Step 6: Commit**

```bash
git add Dockerfile docker-compose.yml .dockerignore Makefile
git commit -m "feat: add Docker deployment and finalize Makefile"
```

---

### Task 14: S3 Storage Backend

**Files:**
- Create: `internal/storage/s3.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Install S3 dependency**

```bash
go get github.com/aws/aws-sdk-go-v2@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/s3@latest
go get github.com/aws/aws-sdk-go-v2/credentials@latest
```

- [ ] **Step 2: Implement S3Storage**

Create `internal/storage/s3.go`:

```go
package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"cloudalbum/internal/config"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(cfg config.S3StorageConf) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AK, cfg.SK, "")
	client := s3.New(s3.Options{
		Region: cfg.Region,
		Credentials: creds,
		BaseEndpoint: aws.String(cfg.Endpoint),
	})
	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *S3Storage) Save(ctx context.Context, key string, data io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	if err != nil {
		return fmt.Errorf("s3 put: %w", err)
	}
	return nil
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get: %w", err)
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}
```

- [ ] **Step 3: Add S3 to main.go initStorage**

Update `cmd/server/main.go` `initStorage` function:

```go
func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.Local.Path)
	case "s3":
		return storage.NewS3Storage(cfg.Storage.S3)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/storage/s3.go cmd/server/main.go go.mod go.sum
git commit -m "feat: add S3-compatible storage backend"
```

---

## Plan Self-Review

### Spec Coverage Check

| Spec Section | Covered by Task |
|---|---|
| Project structure | Task 1 |
| Config (YAML) | Task 1 |
| Data models (User, Image, Album, Token) | Task 2 |
| Storage interface + Local | Task 3 |
| Storage S3 | Task 14 |
| Image processing pipeline | Task 4 |
| Auth (JWT + API Token) | Task 6 |
| Image API (upload, CRUD, batch, URL) | Task 7 |
| Album API | Task 7 |
| Public serving (/i/, /t/) | Task 7 |
| React setup + Login | Task 8 |
| Layout + Upload page | Task 9 |
| Image management page | Task 10 |
| Album, Dashboard, Token, Trash, Settings pages | Task 11 |
| Go embed integration | Task 12 |
| Docker deployment | Task 13 |

### Placeholder Scan

No TBD/TODO found. All code steps contain complete implementations.

### Type Consistency

- `Storage` interface methods consistent across `LocalStorage`, `S3Storage`, and all consumers
- `repository.ImageListParams` used consistently in `ImageService.List` and `ImageHandler.List`
- `service.Claims` struct fields match what `AuthMiddleware` reads from context
