package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"cloudalbum/internal/config"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	svc *service.SettingsService
}

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Snapshot())
}

func (h *SettingsHandler) Update(c *gin.Context) {
	if c.GetString("auth_type") != "jwt" {
		c.JSON(http.StatusForbidden, gin.H{"error": "api_token_forbidden"})
		return
	}

	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if field, ok := containsUnknownField(raw); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_field", "field": field})
		return
	}

	var input config.Overrides
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_field", "detail": err.Error()})
		return
	}

	if err := h.svc.Update(input, c.GetUint("user_id")); err != nil {
		if errors.Is(err, service.ErrInvalidSetting) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_value", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "settings_persist_failed"})
		return
	}
	c.JSON(http.StatusOK, h.svc.Snapshot())
}

// containsUnknownField 检查 raw JSON 顶层 keys 是否全在白名单 {"server","image"} 内，
// 且 server / image 子对象内的 keys 是否全在嵌套白名单内。
func containsUnknownField(raw []byte) (string, bool) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return "_root", false
	}
	allowedTop := map[string]map[string]bool{
		"server": {"base_url": true},
		"image":  {"max_size": true, "allowed_types": true, "auto_convert": true, "quality": true, "strip_exif": true},
	}
	for key, val := range top {
		nested, ok := allowedTop[key]
		if !ok {
			return key, false
		}
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(val, &sub); err != nil {
			return key, false
		}
		for k := range sub {
			if !nested[k] {
				return key + "." + k, false
			}
		}
	}
	return "", true
}
