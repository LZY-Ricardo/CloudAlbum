package config

import "sync/atomic"

// Provider 持有运行时配置快照。读侧零锁，写侧用 atomic.Pointer 整体替换。
type Provider struct {
	base      Config
	snapshot  atomic.Pointer[Config]
	overrides atomic.Pointer[Overrides]
}

func NewProvider(base Config, overrides Overrides) *Provider {
	p := &Provider{base: base}
	p.applyOverrides(overrides)
	return p
}

// Get 返回当前生效的只读快照。调用方按约定不可修改。
func (p *Provider) Get() *Config { return p.snapshot.Load() }

// Overrides 返回当前 overrides 副本，用于 UI 回显。
func (p *Provider) Overrides() *Overrides { return p.overrides.Load() }

// Apply 用新的 overrides 重新构建快照并原子替换。
func (p *Provider) Apply(o Overrides) { p.applyOverrides(o) }

func (p *Provider) applyOverrides(o Overrides) {
	merged := p.base
	if o.Server.BaseURL != nil {
		merged.Server.BaseURL = *o.Server.BaseURL
	}
	if o.Image.MaxSize != nil {
		merged.Image.MaxSize = *o.Image.MaxSize
	}
	if o.Image.AllowedTypes != nil {
		merged.Image.AllowedTypes = *o.Image.AllowedTypes
	}
	if o.Image.AutoConvert != nil {
		merged.Image.AutoConvert = *o.Image.AutoConvert
	}
	if o.Image.Quality != nil {
		merged.Image.Quality = *o.Image.Quality
	}
	if o.Image.StripExif != nil {
		merged.Image.StripExif = *o.Image.StripExif
	}
	p.snapshot.Store(&merged)
	p.overrides.Store(&o)
}
