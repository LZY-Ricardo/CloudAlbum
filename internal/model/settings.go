package model

import "time"

// Settings 表存运行时配置 overrides，固定单行 id=1，payload 为 JSON。
type Settings struct {
	ID        uint   `gorm:"primaryKey"`
	Payload   string `gorm:"type:text;not null"`
	UpdatedAt time.Time
	UpdatedBy uint
}
