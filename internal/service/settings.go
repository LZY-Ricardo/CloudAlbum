package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"cloudalbum/internal/config"
	"cloudalbum/internal/repository"
)

var ErrInvalidSetting = errors.New("invalid setting")

var supportedTypes = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true,
	"webp": true, "bmp": true, "svg": true,
}

// SettingsSnapshot 包含 UI 渲染所需的全部信息：当前生效值、显式 overrides、白名单。
type SettingsSnapshot struct {
	Effective      EffectiveSettings `json:"effective"`
	Overrides      OverrideFlags     `json:"overrides"`
	EditableFields []string          `json:"editable_fields"`
}

type EffectiveSettings struct {
	Server struct {
		BaseURL string `json:"base_url"`
	} `json:"server"`
	Image struct {
		MaxSize      int64    `json:"max_size"`
		AllowedTypes []string `json:"allowed_types"`
		AutoConvert  string   `json:"auto_convert"`
		Quality      int      `json:"quality"`
		StripExif    bool     `json:"strip_exif"`
	} `json:"image"`
}

type OverrideFlags struct {
	Server struct {
		BaseURL bool `json:"base_url,omitempty"`
	} `json:"server"`
	Image struct {
		MaxSize      bool `json:"max_size,omitempty"`
		AllowedTypes bool `json:"allowed_types,omitempty"`
		AutoConvert  bool `json:"auto_convert,omitempty"`
		Quality      bool `json:"quality,omitempty"`
		StripExif    bool `json:"strip_exif,omitempty"`
	} `json:"image"`
}

var editableFields = []string{
	"server.base_url",
	"image.max_size",
	"image.allowed_types",
	"image.auto_convert",
	"image.quality",
	"image.strip_exif",
}

type SettingsService struct {
	repo     *repository.SettingsRepository
	provider *config.Provider
	mu       sync.Mutex
}

func NewSettingsService(repo *repository.SettingsRepository, provider *config.Provider) *SettingsService {
	return &SettingsService{repo: repo, provider: provider}
}

func (s *SettingsService) Snapshot() SettingsSnapshot {
	cfg := s.provider.Get()
	o := s.provider.Overrides()

	snap := SettingsSnapshot{EditableFields: editableFields}
	snap.Effective.Server.BaseURL = cfg.Server.BaseURL
	snap.Effective.Image.MaxSize = cfg.Image.MaxSize
	snap.Effective.Image.AllowedTypes = cfg.Image.AllowedTypes
	snap.Effective.Image.AutoConvert = cfg.Image.AutoConvert
	snap.Effective.Image.Quality = cfg.Image.Quality
	snap.Effective.Image.StripExif = cfg.Image.StripExif
	if o != nil {
		snap.Overrides.Server.BaseURL = o.Server.BaseURL != nil
		snap.Overrides.Image.MaxSize = o.Image.MaxSize != nil
		snap.Overrides.Image.AllowedTypes = o.Image.AllowedTypes != nil
		snap.Overrides.Image.AutoConvert = o.Image.AutoConvert != nil
		snap.Overrides.Image.Quality = o.Image.Quality != nil
		snap.Overrides.Image.StripExif = o.Image.StripExif != nil
	}
	return snap
}

// Update 校验入参 → 与当前 overrides merge → 持久化 → 通知 Provider 应用 → 全程互斥。
func (s *SettingsService) Update(input config.Overrides, userID uint) error {
	if err := validate(input); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.repo.LoadOrBootstrap()
	if err != nil {
		return fmt.Errorf("load current: %w", err)
	}
	merged := mergeOverrides(current, input)
	if err := s.repo.Save(merged, userID); err != nil {
		return fmt.Errorf("persist settings: %w", err)
	}
	s.provider.Apply(merged)
	return nil
}

func mergeOverrides(base, patch config.Overrides) config.Overrides {
	out := base
	if patch.Server.BaseURL != nil {
		out.Server.BaseURL = patch.Server.BaseURL
	}
	if patch.Image.MaxSize != nil {
		out.Image.MaxSize = patch.Image.MaxSize
	}
	if patch.Image.AllowedTypes != nil {
		out.Image.AllowedTypes = patch.Image.AllowedTypes
	}
	if patch.Image.AutoConvert != nil {
		out.Image.AutoConvert = patch.Image.AutoConvert
	}
	if patch.Image.Quality != nil {
		out.Image.Quality = patch.Image.Quality
	}
	if patch.Image.StripExif != nil {
		out.Image.StripExif = patch.Image.StripExif
	}
	return out
}

func validate(o config.Overrides) error {
	if o.Server.BaseURL != nil {
		u, err := url.Parse(*o.Server.BaseURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("%w: server.base_url must be http/https URL", ErrInvalidSetting)
		}
	}
	if o.Image.MaxSize != nil {
		if *o.Image.MaxSize <= 0 || *o.Image.MaxSize > (1<<30) {
			return fmt.Errorf("%w: image.max_size must be in (0, 1 GiB]", ErrInvalidSetting)
		}
	}
	if o.Image.AllowedTypes != nil {
		if len(*o.Image.AllowedTypes) == 0 {
			return fmt.Errorf("%w: image.allowed_types must be non-empty", ErrInvalidSetting)
		}
		for _, raw := range *o.Image.AllowedTypes {
			t := strings.ToLower(strings.TrimSpace(raw))
			if !supportedTypes[t] {
				return fmt.Errorf("%w: image.allowed_types contains unsupported %q", ErrInvalidSetting, raw)
			}
		}
	}
	if o.Image.AutoConvert != nil {
		v := strings.ToLower(*o.Image.AutoConvert)
		if v != "" && v != "webp" && v != "jpeg" && v != "jpg" {
			return fmt.Errorf("%w: image.auto_convert must be one of '', 'webp', 'jpeg'", ErrInvalidSetting)
		}
	}
	if o.Image.Quality != nil {
		if *o.Image.Quality < 1 || *o.Image.Quality > 100 {
			return fmt.Errorf("%w: image.quality must be in [1, 100]", ErrInvalidSetting)
		}
	}
	return nil
}
