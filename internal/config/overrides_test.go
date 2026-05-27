package config

import (
	"encoding/json"
	"testing"
)

func TestOverridesEmptyJSON(t *testing.T) {
	var o Overrides
	if err := json.Unmarshal([]byte(`{}`), &o); err != nil {
		t.Fatalf("unmarshal empty: %v", err)
	}
	if o.Server.BaseURL != nil {
		t.Fatalf("BaseURL should be nil")
	}
	if o.Image.Quality != nil {
		t.Fatalf("Quality should be nil")
	}
}

func TestOverridesPartialJSON(t *testing.T) {
	var o Overrides
	raw := `{"server":{"base_url":"https://x"},"image":{"quality":90,"strip_exif":true}}`
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if o.Server.BaseURL == nil || *o.Server.BaseURL != "https://x" {
		t.Fatalf("BaseURL: %#v", o.Server.BaseURL)
	}
	if o.Image.Quality == nil || *o.Image.Quality != 90 {
		t.Fatalf("Quality: %#v", o.Image.Quality)
	}
	if o.Image.StripExif == nil || *o.Image.StripExif != true {
		t.Fatalf("StripExif: %#v", o.Image.StripExif)
	}
	if o.Image.MaxSize != nil {
		t.Fatalf("MaxSize should remain nil")
	}
}

func TestOverridesMarshalOmitEmpty(t *testing.T) {
	var o Overrides
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `{"server":{},"image":{}}` {
		t.Fatalf("unexpected: %s", data)
	}
}
