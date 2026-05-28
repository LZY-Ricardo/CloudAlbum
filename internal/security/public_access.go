package security

import (
	"net/url"
	"strings"
)

func AllowPublicAccess(mode string, allowedHosts []string, referer string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "off":
		return true
	case "referer_whitelist":
		host, ok := refererHost(referer)
		if !ok {
			return false
		}
		return hostAllowed(host, allowedHosts)
	case "allow_empty_or_whitelist":
		if strings.TrimSpace(referer) == "" {
			return true
		}
		host, ok := refererHost(referer)
		if !ok {
			return false
		}
		return hostAllowed(host, allowedHosts)
	default:
		return false
	}
}

func refererHost(referer string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(referer))
	if err != nil {
		return "", false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return "", false
	}
	return host, true
}

func hostAllowed(host string, allowedHosts []string) bool {
	for _, allowed := range allowedHosts {
		if strings.ToLower(strings.TrimSpace(allowed)) == host {
			return true
		}
	}
	return false
}
