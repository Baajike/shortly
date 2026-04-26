package utils

import (
	"errors"
	"net"
	"net/url"
)

var (
	ErrInvalidURL    = errors.New("invalid URL format")
	ErrMissingScheme = errors.New("URL must begin with http:// or https://")
	ErrPrivateIP     = errors.New("URLs pointing to private or loopback addresses are not allowed")
	ErrEmptyHost     = errors.New("URL must contain a valid hostname")
)

// ValidateURL returns nil when rawURL is a well-formed, publicly routable URL.
func ValidateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrMissingScheme
	}

	host := u.Hostname()
	if host == "" {
		return ErrEmptyHost
	}

	// Block literal IP addresses that are private/loopback.
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() {
			return ErrPrivateIP
		}
	}

	// Block localhost by hostname.
	if host == "localhost" {
		return ErrPrivateIP
	}

	return nil
}
