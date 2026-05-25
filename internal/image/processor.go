package image

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

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
	cfg config.ImageConfig
}

func NewProcessor(cfg config.ImageConfig) *Processor {
	return &Processor{cfg: cfg}
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

	encFormat := mimeTypeToImaging(mimeType)

	result.Thumbnails = make(map[string][]byte)
	for _, size := range p.cfg.Thumbnails {
		thumb := imaging.Thumbnail(img, size.Width, size.Height, imaging.Lanczos)
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, thumb, encFormat, imaging.JPEGQuality(p.cfg.Quality)); err != nil {
			return nil, fmt.Errorf("encode thumbnail %s: %w", size.Name, err)
		}
		result.Thumbnails[size.Name] = buf.Bytes()
	}

	return result, nil
}

func (p *Processor) GenerateThumbnailKey(originalKey, sizeName string) string {
	return fmt.Sprintf("thumbs/%s_%s", sizeName, originalKey)
}

func mimeTypeToImaging(mimeType string) imaging.Format {
	switch mimeType {
	case "image/png":
		return imaging.PNG
	case "image/gif":
		return imaging.GIF
	default:
		return imaging.JPEG
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
	if len(data) < 512 {
		return ""
	}
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "image/png"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}):
		return "image/gif"
	case len(data) > 11 && string(data[8:12]) == "webp":
		return "image/webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return "image/bmp"
	case bytes.HasPrefix(data, []byte("<?xml")) || bytes.HasPrefix(data, []byte("<svg")):
		return "image/svg+xml"
	}
	return ""
}
