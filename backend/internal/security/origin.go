package security

import (
	"net/url"
	"strings"
)

// ValidateOrigin checks whether the incoming Origin header is permitted.
// If allowedOrigins contains "*", all origins are accepted.
// If the origin header is empty (e.g., CLI tools or native apps), it is accepted.
func ValidateOrigin(originHeader string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}

	for _, pattern := range allowedOrigins {
		if pattern == "*" {
			return true
		}
	}

	origin := strings.TrimSpace(originHeader)
	if origin == "" {
		return true
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false
	}

	originHost := strings.ToLower(parsedOrigin.Host)
	for _, pattern := range allowedOrigins {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Exact match
		if strings.EqualFold(pattern, origin) {
			return true
		}

		// Parse pattern as URL or host
		parsedPattern, err := url.Parse(pattern)
		if err == nil && parsedPattern.Host != "" {
			if strings.EqualFold(parsedPattern.Host, originHost) {
				return true
			}
		} else {
			if strings.EqualFold(pattern, originHost) {
				return true
			}
		}
	}

	return false
}
