package model

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	StorageKey   string         `gorm:"index;size:255;not null" json:"storage_key"`
	Filename     string         `gorm:"size:255;not null" json:"filename"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	Size         int64          `json:"size"`
	MimeType     string         `gorm:"size:100" json:"mime_type"`
	Width        int            `json:"width"`
	Height       int            `json:"height"`
	Hash         string         `gorm:"index;size:64" json:"hash"`
	AlbumID      *uint          `gorm:"index" json:"album_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
	Album        *Album         `gorm:"foreignKey:AlbumID" json:"-"`
}
