package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/disintegration/imaging"

	"cloudalbum/internal/config"
)

type ProcessResult struct {
	Width      int
	Height     int
	Hash       string
	Size       int64
	MimeType   string
	Thumbnails map[string][]byte
}

type Processor struct {
	provider *config.Provider
}

func NewProcessor(provider *config.Provider) *Processor {
	return &Processor{provider: provider}
}

func (p *Processor) imageCfg() config.ImageConfig {
	return p.provider.Get().Image
}

func (p *Processor) Process(data []byte, mimeType string) (*ProcessResult, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	result := &ProcessResult{
		Width:    bounds.Dx(),
		Height:   bounds.Dy(),
		Size:     int64(len(data)),
		MimeType: mimeType,
	}

	hash := sha256.Sum256(data)
	result.Hash = hex.EncodeToString(hash[:])

	cfg := p.imageCfg()
	thumbFormat, thumbOptions := p.thumbnailEncoding()

	result.Thumbnails = make(map[string][]byte)
	for _, size := range cfg.Thumbnails {
		thumb := imaging.Thumbnail(img, size.Width, size.Height, imaging.Lanczos)
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, thumb, thumbFormat, thumbOptions...); err != nil {
			return nil, fmt.Errorf("encode thumbnail %s: %w", size.Name, err)
		}
		result.Thumbnails[size.Name] = buf.Bytes()
	}

	return result, nil
}

func (p *Processor) GenerateThumbnailKey(originalKey, sizeName string) string {
	return fmt.Sprintf("thumbs/%s_%s", sizeName, originalKey)
}

func (p *Processor) thumbnailEncoding() (imaging.Format, []imaging.EncodeOption) {
	cfg := p.imageCfg()
	switch strings.ToLower(cfg.AutoConvert) {
	case "", "jpg", "jpeg", "webp":
		return imaging.JPEG, []imaging.EncodeOption{imaging.JPEGQuality(cfg.Quality)}
	case "png":
		return imaging.PNG, nil
	case "gif":
		return imaging.GIF, nil
	default:
		return imaging.JPEG, []imaging.EncodeOption{imaging.JPEGQuality(cfg.Quality)}
	}
}

func HashFromReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func DetectImageType(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "image/png"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}):
		return "image/gif"
	case hasWebPSignature(data):
		return "image/webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return "image/bmp"
	case looksLikeSVG(data):
		return "image/svg+xml"
	}
	return ""
}

func hasWebPSignature(data []byte) bool {
	return len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && string(data[8:12]) == "WEBP"
}

func looksLikeSVG(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false
	}

	lower := strings.ToLower(string(trimmed))
	if strings.HasPrefix(lower, "<svg") {
		return true
	}
	if strings.HasPrefix(lower, "<?xml") {
		end := strings.Index(lower, "?>")
		if end == -1 {
			return false
		}
		rest := strings.TrimSpace(lower[end+2:])
		return strings.HasPrefix(rest, "<svg")
	}

	return false
}
