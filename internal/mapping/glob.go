package mapping

import (
	"strings"

	"github.com/gobwas/glob"
)

// ValidateGlobPattern checks if a pattern is a valid glob pattern.
func ValidateGlobPattern(pattern string) error {
	_, err := glob.Compile(strings.ToLower(pattern))
	return err
}

// SuggestPattern creates a suggested glob pattern from a value.
// It wraps the value in wildcards for common merchant name patterns.
func SuggestPattern(value string) string {
	// If the value already contains wildcards, return as-is
	if containsGlobChars(value) {
		return value
	}

	// For typical merchant names, suggest wrapping with wildcards
	return "*" + value + "*"
}

// IsExactPattern returns true if the pattern has no glob characters.
func IsExactPattern(pattern string) bool {
	return !containsGlobChars(pattern)
}
