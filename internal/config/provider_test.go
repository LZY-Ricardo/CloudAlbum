package config

import (
	"sync"
	"testing"
)

func baseFixture() Config {
	return Config{
		Server: ServerConfig{Port: 8080, BaseURL: "http://localhost:8080"},
		Image: ImageConfig{
			MaxSize:      50 << 20,
			AllowedTypes: []string{"jpg", "png"},
			AutoConvert:  "webp",
			Quality:      85,
			StripExif:    true,
		},
		Auth: AuthConfig{JWTSecret: "s", TokenExpire: 168 * 3600 * 1e9},
	}
}

func TestProviderEmptyOverridesEqualsBase(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	got := p.Get()
	if got.Server.BaseURL != "http://localhost:8080" {
		t.Fatalf("BaseURL: %s", got.Server.BaseURL)
	}
	if got.Image.Quality != 85 {
		t.Fatalf("Quality: %d", got.Image.Quality)
	}
}

func TestProviderApplyOverrides(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	var o Overrides
	url := "https://img.example.com"
	q := 90
	o.Server.BaseURL = &url
	o.Image.Quality = &q
	p.Apply(o)

	got := p.Get()
	if got.Server.BaseURL != url {
		t.Fatalf("BaseURL not overridden: %s", got.Server.BaseURL)
	}
	if got.Image.Quality != 90 {
		t.Fatalf("Quality not overridden: %d", got.Image.Quality)
	}
	if got.Image.MaxSize != 50<<20 {
		t.Fatalf("MaxSize should keep base: %d", got.Image.MaxSize)
	}
}

func TestProviderConcurrentReadWrite(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = p.Get()
				}
			}
		}()
	}
	for i := 0; i < 200; i++ {
		var o Overrides
		q := i%100 + 1
		o.Image.Quality = &q
		p.Apply(o)
	}
	close(stop)
	wg.Wait()
}
