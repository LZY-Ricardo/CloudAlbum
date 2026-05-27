package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"cloudalbum/internal/config"
)

func TestDetectImageType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF}, want: "image/jpeg"},
		{name: "png", data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, want: "image/png"},
		{name: "gif", data: []byte("GIF89a"), want: "image/gif"},
		{name: "bmp", data: []byte("BMrest"), want: "image/bmp"},
		{name: "valid webp", data: []byte("RIFF1234WEBPVP8 "), want: "image/webp"},
		{name: "webp without riff", data: []byte("NOPE1234WEBPVP8 "), want: ""},
		{name: "webp without webp marker", data: []byte("RIFF1234NOPEVP8 "), want: ""},
		{name: "svg tag", data: []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"), want: "image/svg+xml"},
		{name: "svg after xml declaration", data: []byte("<?xml version=\"1.0\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"), want: "image/svg+xml"},
		{name: "plain xml is not svg", data: []byte("<?xml version=\"1.0\"?><note></note>"), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectImageType(tt.data); got != tt.want {
				t.Fatalf("DetectImageType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessUsesDeterministicJPEGThumbnailsAndAppliesQuality(t *testing.T) {
	data := createPNGTestImage(t)
	thumbnails := []config.ThumbnailSize{{Name: "thumb", Width: 128, Height: 128}}

	highQuality := NewProcessor(newProcessorTestProvider(config.ImageConfig{
		AutoConvert: "webp",
		Quality:     95,
		Thumbnails:  thumbnails,
	}))
	lowQuality := NewProcessor(newProcessorTestProvider(config.ImageConfig{
		AutoConvert: "webp",
		Quality:     25,
		Thumbnails:  thumbnails,
	}))

	highResult, err := highQuality.Process(data, "image/png")
	if err != nil {
		t.Fatalf("highQuality.Process() error = %v", err)
	}
	lowResult, err := lowQuality.Process(data, "image/png")
	if err != nil {
		t.Fatalf("lowQuality.Process() error = %v", err)
	}

	highThumb := highResult.Thumbnails["thumb"]
	lowThumb := lowResult.Thumbnails["thumb"]

	if !bytes.HasPrefix(highThumb, []byte{0xFF, 0xD8}) {
		t.Fatalf("high quality thumbnail is not JPEG: % x", highThumb[:min(len(highThumb), 4)])
	}
	if !bytes.HasPrefix(lowThumb, []byte{0xFF, 0xD8}) {
		t.Fatalf("low quality thumbnail is not JPEG: % x", lowThumb[:min(len(lowThumb), 4)])
	}
	if len(highThumb) == len(lowThumb) {
		t.Fatalf("thumbnail sizes are identical (%d); expected quality to affect JPEG output", len(highThumb))
	}
}

func createPNGTestImage(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x*17 + y*3) % 256),
				G: uint8((x*7 + y*29) % 256),
				B: uint8((x*y + x + y) % 256),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}
	return buf.Bytes()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func newProcessorTestProvider(img config.ImageConfig) *config.Provider {
	base := config.Config{Image: img}
	return config.NewProvider(base, config.Overrides{})
}
