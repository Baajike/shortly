package utils

import "strings"

// ParsedUA holds the structured fields extracted from a raw User-Agent string.
type ParsedUA struct {
	DeviceType string // mobile | tablet | desktop | bot
	Browser    string
	OS         string
}

// ParseUserAgent performs a lightweight, dependency-free parse of a UA string.
// It is intentionally simple: correctness for the top 95% of traffic, not 100%.
func ParseUserAgent(ua string) ParsedUA {
	lower := strings.ToLower(ua)
	return ParsedUA{
		DeviceType: detectDevice(lower),
		Browser:    detectBrowser(lower),
		OS:         detectOS(lower),
	}
}

func detectDevice(ua string) string {
	for _, kw := range []string{"bot", "crawler", "spider", "slurp", "googlebot", "bingbot", "yandex", "baidu", "duckduckbot"} {
		if strings.Contains(ua, kw) {
			return "bot"
		}
	}
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") || strings.Contains(ua, "kindle") {
		return "tablet"
	}
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") || strings.Contains(ua, "android") {
		return "mobile"
	}
	return "desktop"
}

func detectBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/"):
		return "Edge"
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "firefox") || strings.Contains(ua, "fxios"):
		return "Firefox"
	case strings.Contains(ua, "samsungbrowser"):
		return "Samsung"
	case strings.Contains(ua, "chrome") || strings.Contains(ua, "crios"):
		return "Chrome"
	case strings.Contains(ua, "safari"):
		return "Safari"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident/"):
		return "IE"
	default:
		return "Other"
	}
}

func detectOS(ua string) string {
	switch {
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macos"):
		return "macOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Other"
	}
}
