package utils

import "testing"

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trim and lowercase", input: "  User@Example.COM ", want: "user@example.com"},
		{name: "invalid", input: "not-an-email", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeEmail(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("NormalizeEmail() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("NormalizeEmail() = %q, want %q", got, test.want)
			}
		})
	}
}
