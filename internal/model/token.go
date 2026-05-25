package model

import "time"

type APIToken struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	Name       string     `gorm:"size:100;not null" json:"name"`
	TokenHash  string     `gorm:"uniqueIndex;size:64;not null" json:"-"`
	Scope      string     `gorm:"size:20;default:full" json:"scope"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
