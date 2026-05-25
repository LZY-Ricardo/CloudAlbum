package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	imgpkg "cloudalbum/internal/image"
	"cloudalbum/internal/storage"
	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	store     storage.Storage
	processor *imgpkg.Processor
}

func NewPublicHandler(store storage.Storage, processor *imgpkg.Processor) *PublicHandler {
	return &PublicHandler{store: store, processor: processor}
}

func (h *PublicHandler) Image(c *gin.Context) {
	h.serve(c, strings.TrimPrefix(c.Param("key"), "/"))
}

func (h *PublicHandler) Thumbnail(c *gin.Context) {
	key := strings.TrimPrefix(c.Param("key"), "/")
	if key == "" {
		c.Status(http.StatusNotFound)
		return
	}
	h.serve(c, "thumbs/"+key)
}

func (h *PublicHandler) serve(c *gin.Context, key string) {
	if strings.TrimSpace(key) == "" {
		c.Status(http.StatusNotFound)
		return
	}

	reader, err := h.store.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || strings.Contains(strings.ToLower(err.Error()), "not found") {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	if contentType := extToContentType(filepath.Ext(key)); contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, max-age=31536000")
	_, _ = io.Copy(c.Writer, reader)
}

func extToContentType(ext string) string {
	switch strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".") {
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
