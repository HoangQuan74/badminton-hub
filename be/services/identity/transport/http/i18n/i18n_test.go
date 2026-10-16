package i18n

import "testing"

func TestTranslate(t *testing.T) {
	service, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	tests := []struct{ language, key, want string }{
		{language: "", key: "auth.invalid_credentials", want: "Email or password is incorrect."},
		{language: "en", key: "auth.invalid_credentials", want: "Email or password is incorrect."},
		{language: "vi", key: "auth.invalid_credentials", want: "Email hoặc mật khẩu không đúng."},
		{language: "vi-VN,vi;q=0.9,en;q=0.8", key: "auth.email_already_registered", want: "Email này đã được đăng ký."},
	}
	for _, test := range tests {
		got, err := service.Translate(test.language, test.key)
		if err != nil {
			t.Fatalf("Translate(%q, %q) error = %v", test.language, test.key, err)
		}
		if got != test.want {
			t.Fatalf("Translate(%q, %q) = %q, want %q", test.language, test.key, got, test.want)
		}
	}
}
