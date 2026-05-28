package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudalbum/internal/config"
	imgpkg "cloudalbum/internal/image"
	"github.com/gin-gonic/gin"
)

type stubPublicStorage struct {
	files map[string]string
}

func (s *stubPublicStorage) Save(_ context.Context, _ string, _ io.Reader) error { return nil }
func (s *stubPublicStorage) Delete(_ context.Context, _ string) error            { return nil }
func (s *stubPublicStorage) Exists(_ context.Context, key string) (bool, error) {
	_, ok := s.files[key]
	return ok, nil
}
func (s *stubPublicStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	content, ok := s.files[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func TestPublicHandlerImageServesStoredFileWithHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(rec)
	provider := testPublicProvider()
	h := NewPublicHandler(&stubPublicStorage{files: map[string]string{"demo/test.jpg": "image-bytes"}}, imgpkg.NewProcessor(provider), provider)
	engine.GET("/i/*key", h.Image)

	req := httptest.NewRequest(http.MethodGet, "/i/demo/test.jpg", nil)
	engine.ServeHTTP(rec, req)
	_ = ctx

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("content-type = %q, want image/jpeg", got)
	}
	if got := rec.Header().Get("Cache-Control"); got == "" {
		t.Fatal("cache-control header missing")
	}
	if got := rec.Header().Get("Vary"); !strings.Contains(got, "Referer") {
		t.Fatalf("vary = %q, want Referer", got)
	}
	if rec.Body.String() != "image-bytes" {
		t.Fatalf("body = %q, want image-bytes", rec.Body.String())
	}
}

func TestPublicHandlerImageReturnsNotFoundWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	provider := testPublicProvider()
	h := NewPublicHandler(&stubPublicStorage{files: map[string]string{}}, imgpkg.NewProcessor(provider), provider)
	engine.GET("/i/*key", h.Image)

	req := httptest.NewRequest(http.MethodGet, "/i/missing/test.jpg", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func testPublicImageConfig() config.ImageConfig {
	return config.ImageConfig{AutoConvert: "jpeg", Quality: 85}
}

func TestPublicHandlerRejectsDisallowedReferer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	provider := config.NewProvider(config.Config{
		Image:        testPublicImageConfig(),
		PublicAccess: config.PublicAccessConfig{Mode: "referer_whitelist", AllowedRefererHosts: []string{"good.example"}},
	}, config.Overrides{})
	h := NewPublicHandler(&stubPublicStorage{files: map[string]string{"demo/test.jpg": "image-bytes"}}, imgpkg.NewProcessor(provider), provider)
	engine.GET("/i/*key", h.Image)

	req := httptest.NewRequest(http.MethodGet, "/i/demo/test.jpg", nil)
	req.Header.Set("Referer", "https://evil.example/path")
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "public_access_forbidden") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func testPublicProvider() *config.Provider {
	base := config.Config{Image: testPublicImageConfig()}
	return config.NewProvider(base, config.Overrides{})
}
