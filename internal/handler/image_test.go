package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudalbum/internal/ratelimit"
	"github.com/gin-gonic/gin"
)

func TestImageUploadReturns429WhenRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	limiter := ratelimit.NewLimiter(true, time.Minute, 1)
	if err := limiter.Allow("jwt:user:1"); err != nil {
		t.Fatalf("preload limiter: %v", err)
	}
	h := &ImageHandler{uploadLimiter: limiter}
	engine.POST("/images", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("auth_type", "jwt")
		h.Upload(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "a.jpg")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("abc")); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "{\"error\":\"rate_limit_exceeded\"}" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
