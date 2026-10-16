package utils

import (
	"fmt"
	"net/mail"
	"strings"
)

func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", fmt.Errorf("invalid email")
	}
	return email, nil
}
