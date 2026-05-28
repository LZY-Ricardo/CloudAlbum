package security

import "testing"

func TestAllowPublicAccessModes(t *testing.T) {
	if !AllowPublicAccess("off", nil, "") {
		t.Fatal("off mode should allow")
	}
	if AllowPublicAccess("referer_whitelist", []string{"example.com"}, "") {
		t.Fatal("whitelist mode should reject empty referer")
	}
	if !AllowPublicAccess("allow_empty_or_whitelist", []string{"example.com"}, "") {
		t.Fatal("allow_empty_or_whitelist should allow empty referer")
	}
	if !AllowPublicAccess("referer_whitelist", []string{"example.com"}, "https://example.com/a.jpg") {
		t.Fatal("expected allowed host")
	}
	if AllowPublicAccess("referer_whitelist", []string{"example.com"}, "https://evil.com/a.jpg") {
		t.Fatal("unexpected allowed host")
	}
}
