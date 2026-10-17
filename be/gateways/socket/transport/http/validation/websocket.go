package validation

import "strings"

const MaxMessageSize = 1 << 20

func AccessToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if len(token) == 0 || len(token) > 4096 || strings.Count(token, ".") != 2 {
		return "", false
	}
	return token, true
}
