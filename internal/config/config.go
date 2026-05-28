package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Database        DatabaseConfig        `yaml:"database"`
	Storage         StorageConfig         `yaml:"storage"`
	Image           ImageConfig           `yaml:"image"`
	Auth            AuthConfig            `yaml:"auth"`
	Token           TokenPolicyConfig     `yaml:"token"`
	UploadRateLimit UploadRateLimitConfig `yaml:"upload_rate_limit"`
	PublicAccess    PublicAccessConfig    `yaml:"public_access"`
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
	MaxSize      int64           `yaml:"max_size"`
	AllowedTypes []string        `yaml:"allowed_types"`
	AutoConvert  string          `yaml:"auto_convert"`
	Quality      int             `yaml:"quality"`
	StripExif    bool            `yaml:"strip_exif"`
	Thumbnails   []ThumbnailSize `yaml:"thumbnails"`
}

type AuthConfig struct {
	JWTSecret   string        `yaml:"jwt_secret"`
	TokenExpire time.Duration `yaml:"token_expire"`
}

type TokenPolicyConfig struct {
	AllowNoExpiry    bool          `yaml:"allow_no_expiry"`
	DefaultExpiresIn time.Duration `yaml:"default_expires_in"`
}

type UploadRateLimitConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Window      time.Duration `yaml:"window"`
	MaxRequests int           `yaml:"max_requests"`
}

type PublicAccessConfig struct {
	Mode                string   `yaml:"mode"`
	AllowedRefererHosts []string `yaml:"allowed_referer_hosts"`
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
		cfg.Image.MaxSize = 50 << 20
	}
	if cfg.Image.Quality == 0 {
		cfg.Image.Quality = 85
	}
	if cfg.Auth.TokenExpire == 0 {
		cfg.Auth.TokenExpire = 7 * 24 * time.Hour
	}
	if cfg.Token.DefaultExpiresIn == 0 {
		cfg.Token.DefaultExpiresIn = 7 * 24 * time.Hour
	}
	cfg.Token.AllowNoExpiry = true
	if cfg.UploadRateLimit.Window == 0 {
		cfg.UploadRateLimit.Window = time.Minute
	}
	if cfg.UploadRateLimit.MaxRequests == 0 {
		cfg.UploadRateLimit.MaxRequests = 20
	}
	if cfg.PublicAccess.Mode == "" {
		cfg.PublicAccess.Mode = "off"
	}
	if cfg.Server.BaseURL == "" {
		cfg.Server.BaseURL = "http://localhost:8080"
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "change-me-in-production"
	}
	if len(cfg.Image.AllowedTypes) == 0 {
		cfg.Image.AllowedTypes = []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"}
	}
	if cfg.Image.AutoConvert == "" {
		cfg.Image.AutoConvert = "webp"
	}
	if len(cfg.Image.Thumbnails) == 0 {
		cfg.Image.Thumbnails = []ThumbnailSize{
			{Name: "thumb", Width: 200, Height: 200},
			{Name: "medium", Width: 800, Height: 600},
			{Name: "large", Width: 1200, Height: 900},
		}
	}
	return &cfg, nil
}
