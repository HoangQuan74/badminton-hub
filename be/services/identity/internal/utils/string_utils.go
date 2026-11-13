package utils

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NormalizeLowercase canonicalizes case-insensitive text without validating it.
func NormalizeLowercase(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// IsRequiredText checks normalized text against a caller-owned rune limit.
func IsRequiredText(value string, maxRunes int) bool {
	return value != "" && utf8.RuneCountInString(value) <= maxRunes
}

// IsSlug checks normalized ASCII slug syntax and a caller-owned length limit.
func IsSlug(value string, maxLength int) bool {
	return len(value) <= maxLength && slugPattern.MatchString(value)
}
