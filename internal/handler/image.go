package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cloudalbum/internal/ratelimit"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ImageHandler struct {
	imageSvc      *service.ImageService
	uploadLimiter *ratelimit.Limiter
}

func NewImageHandler(imageSvc *service.ImageService, uploadLimiter *ratelimit.Limiter) *ImageHandler {
	return &ImageHandler{imageSvc: imageSvc, uploadLimiter: uploadLimiter}
}

func (h *ImageHandler) Upload(c *gin.Context) {
	if h.limitUpload(c) {
		return
	}

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

	albumID, err := optionalUintFromString(c.PostForm("album_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid album_id"})
		return
	}

	results := make([]gin.H, 0, len(files))
	userID := c.GetUint("user_id")
	for _, file := range files {
		img, err := h.imageSvc.Upload(userID, file, albumID)
		if err != nil {
			results = append(results, gin.H{"filename": file.Filename, "error": err.Error()})
			continue
		}
		results = append(results, gin.H{"image": img, "urls": h.imageSvc.URLs(img)})
	}
	c.JSON(http.StatusCreated, gin.H{"results": results})
}

func (h *ImageHandler) UploadURL(c *gin.Context) {
	if h.limitUpload(c) {
		return
	}

	var req struct {
		URL     string `json:"url" binding:"required"`
		AlbumID *uint  `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	img, err := h.imageSvc.UploadFromURL(c.GetUint("user_id"), req.URL, req.AlbumID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"image": img, "urls": h.imageSvc.URLs(img)})
}

func (h *ImageHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	albumID, err := optionalUintFromString(c.Query("album_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid album_id"})
		return
	}

	images, total, err := h.imageSvc.List(repository.ImageListParams{
		UserID:      c.GetUint("user_id"),
		AlbumID:     albumID,
		Page:        page,
		PageSize:    pageSize,
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		OnlyDeleted: c.Query("deleted") == "true",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, 0, len(images))
	for _, img := range images {
		imageCopy := img
		response = append(response, gin.H{"image": imageCopy, "urls": h.imageSvc.URLs(&imageCopy)})
	}
	c.JSON(http.StatusOK, gin.H{"images": response, "total": total, "page": page, "page_size": pageSize})
}

func (h *ImageHandler) Get(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	img, err := h.imageSvc.Get(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	if img.UserID != c.GetUint("user_id") {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": img, "urls": h.imageSvc.URLs(img)})
}

func (h *ImageHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.imageSvc.Update(id, c.GetUint("user_id"), req); err != nil {
		h.writeImageServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.imageSvc.Delete(id, c.GetUint("user_id")); err != nil {
		h.writeImageServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ImageHandler) Batch(c *gin.Context) {
	var req struct {
		IDs     []uint `json:"ids" binding:"required"`
		Action  string `json:"action" binding:"required"`
		AlbumID *uint  `json:"album_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.imageSvc.BatchOperation(req.IDs, c.GetUint("user_id"), req.Action, req.AlbumID); err != nil {
		h.writeImageServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "done"})
}

func (h *ImageHandler) Restore(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.imageSvc.Restore(id, c.GetUint("user_id")); err != nil {
		h.writeImageServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restored"})
}

func (h *ImageHandler) HardDelete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.imageSvc.HardDelete(id, c.GetUint("user_id")); err != nil {
		h.writeImageServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permanently deleted"})
}

func (h *ImageHandler) Stats(c *gin.Context) {
	count, totalSize, err := h.imageSvc.Stats(c.GetUint("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count, "total_size": totalSize})
}

func (h *ImageHandler) writeImageServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
	case errors.Is(err, service.ErrImageForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func (h *ImageHandler) limitUpload(c *gin.Context) bool {
	if h.uploadLimiter == nil {
		return false
	}
	if err := h.uploadLimiter.Allow(uploadLimitKey(c)); err != nil {
		if errors.Is(err, ratelimit.ErrRateLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limit_exceeded"})
			return true
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}
	return false
}

func uploadLimitKey(c *gin.Context) string {
	if c.GetString("auth_type") == "jwt" {
		return "jwt:user:" + strconv.FormatUint(uint64(c.GetUint("user_id")), 10)
	}
	if tokenID, ok := c.Get("token_id"); ok {
		switch v := tokenID.(type) {
		case uint:
			return "token:" + strconv.FormatUint(uint64(v), 10)
		case uint64:
			return "token:" + strconv.FormatUint(v, 10)
		}
	}
	return "user:" + strconv.FormatUint(uint64(c.GetUint("user_id")), 10)
}

func optionalUintFromString(raw string) (*uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	converted := uint(value)
	return &converted, nil
}
