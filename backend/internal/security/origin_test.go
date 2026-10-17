package security

import (
	"testing"
)

func TestValidateOrigin(t *testing.T) {
	tests := []struct {
		origin   string
		allowed  []string
		expected bool
	}{
		{"https://example.com", []string{"*"}, true},
		{"", []string{"https://example.com"}, true}, // empty origin (e.g. CLI/curl) allowed
		{"https://example.com", []string{"https://example.com"}, true},
		{"https://EXAMPLE.COM", []string{"https://example.com"}, true},
		{"https://evil.com", []string{"https://example.com"}, false},
		{"http://localhost:3000", []string{"http://localhost:3000", "https://syncraft.dev"}, true},
		{"https://syncraft.dev", []string{"http://localhost:3000", "https://syncraft.dev"}, true},
		{"https://attacker.com", []string{"http://localhost:3000", "https://syncraft.dev"}, false},
	}

	for _, tt := range tests {
		result := ValidateOrigin(tt.origin, tt.allowed)
		if result != tt.expected {
			t.Errorf("origin %q with allowed %v: expected %v, got %v", tt.origin, tt.allowed, tt.expected, result)
		}
	}
}
