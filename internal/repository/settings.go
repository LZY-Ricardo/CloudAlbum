package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"

	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// LoadOrBootstrap 读取 settings 表的唯一一行；若表空则插入空 overrides 并返回零值。
// 若 payload 损坏，返回错误，由上层决定回退策略。
func (r *SettingsRepository) LoadOrBootstrap() (config.Overrides, error) {
	var row model.Settings
	err := r.db.First(&row, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		seed := model.Settings{ID: 1, Payload: "{}", UpdatedAt: time.Now()}
		if err := r.db.Create(&seed).Error; err != nil {
			return config.Overrides{}, fmt.Errorf("bootstrap settings: %w", err)
		}
		return config.Overrides{}, nil
	}
	if err != nil {
		return config.Overrides{}, fmt.Errorf("load settings: %w", err)
	}
	var o config.Overrides
	if err := json.Unmarshal([]byte(row.Payload), &o); err != nil {
		return config.Overrides{}, fmt.Errorf("decode settings payload: %w", err)
	}
	return o, nil
}

// Save 用新 overrides 整体覆盖唯一一行；调用方需保证传入的 overrides 已与旧值 merge。
func (r *SettingsRepository) Save(o config.Overrides, updatedBy uint) error {
	data, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return r.db.Model(&model.Settings{}).
		Where("id = ?", 1).
		Updates(map[string]any{
			"payload":    string(data),
			"updated_at": time.Now(),
			"updated_by": updatedBy,
		}).Error
}

// UpdatedAt 返回当前 settings 行的更新时间，用于 yaml mtime 比对。
func (r *SettingsRepository) UpdatedAt() (time.Time, error) {
	var row model.Settings
	if err := r.db.Select("updated_at").First(&row, 1).Error; err != nil {
		return time.Time{}, err
	}
	return row.UpdatedAt, nil
}
