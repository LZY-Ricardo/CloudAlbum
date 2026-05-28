package config

// Overrides 表示运行时对 YAML 基线的覆盖值。
// 所有字段使用指针：nil 表示"未设置 → 落回 YAML 默认"。
type Overrides struct {
	Server struct {
		BaseURL *string `json:"base_url,omitempty"`
	} `json:"server"`
	Image struct {
		MaxSize      *int64    `json:"max_size,omitempty"`
		AllowedTypes *[]string `json:"allowed_types,omitempty"`
		AutoConvert  *string   `json:"auto_convert,omitempty"`
		Quality      *int      `json:"quality,omitempty"`
		StripExif    *bool     `json:"strip_exif,omitempty"`
	} `json:"image"`
}
