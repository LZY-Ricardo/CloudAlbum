package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAppliesSecurityDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("server: {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	if cfg.Token.DefaultExpiresIn != 7*24*time.Hour {
		t.Fatalf("DefaultExpiresIn = %v", cfg.Token.DefaultExpiresIn)
	}
	if cfg.Token.AllowNoExpiry != true {
		t.Fatalf("AllowNoExpiry = %v, want true", cfg.Token.AllowNoExpiry)
	}
	if cfg.UploadRateLimit.Window != time.Minute {
		t.Fatalf("Window = %v, want 1m", cfg.UploadRateLimit.Window)
	}
	if cfg.UploadRateLimit.MaxRequests != 20 {
		t.Fatalf("MaxRequests = %d, want 20", cfg.UploadRateLimit.MaxRequests)
	}
	if cfg.PublicAccess.Mode != "off" {
		t.Fatalf("Mode = %q, want off", cfg.PublicAccess.Mode)
	}
}
