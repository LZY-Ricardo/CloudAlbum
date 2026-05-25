package model

import "time"

type Album struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Description  string    `gorm:"size:500" json:"description"`
	CoverImageID *uint     `json:"cover_image_id"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`

	User       User   `gorm:"foreignKey:UserID" json:"-"`
	CoverImage *Image `gorm:"foreignKey:CoverImageID" json:"-"`
}
