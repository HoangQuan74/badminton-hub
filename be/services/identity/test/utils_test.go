package identitytest

import (
	"testing"

	"github.com/kainguyen/badminton-hub/be/services/identity/internal/utils"
)

func TestTextNormalizationAndLimits(t *testing.T) {
	if value := utils.NormalizeLowercase("\u00a0Owner@Example.TEST\t"); value != "owner@example.test" {
		t.Fatalf("unexpected normalized address: %q", value)
	}

	// Human names are bounded by characters rather than UTF-8 bytes.
	if !utils.IsRequiredText("Nguyễn", 6) || utils.IsRequiredText("Nguyễn", 5) {
		t.Fatal("text length must count Unicode characters")
	}

	if utils.IsRequiredText("", 120) {
		t.Fatal("required text must reject the empty value")
	}

	cases := []struct {
		value string
		limit int
		valid bool
	}{
		{value: "venue-123", limit: 9, valid: true},
		{value: "venue-123", limit: 8},
		{value: "Venue-123", limit: 100},
		{value: "venue--123", limit: 100},
		{value: "-venue", limit: 100},
		{value: "venue-", limit: 100},
		{value: "sân", limit: 100},
		{value: "", limit: 100},
	}
	for _, test := range cases {
		t.Run(test.value, func(t *testing.T) {
			if valid := utils.IsSlug(test.value, test.limit); valid != test.valid {
				t.Fatalf("IsSlug(%q, %d) = %v, want %v", test.value, test.limit, valid, test.valid)
			}
		})
	}
}
