package services

import (
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		// Valid cases
		{"test@example.com", true},
		{"user.name+tag+sorting@example.com", true},
		{"user@sub.example.com", true},
		{"a@b.co", true},

		// Invalid cases
		{"", false},
		{"   ", false},
		{"plainaddress", false},
		{"#@%^%#$@#$@#.com", false},
		{"@example.com", false},
		{"Joe Smith <email@example.com>", false}, // ParseAddress would parse this, but our fn rejects it to ensure exact email address
		{"email.example.com", false},
		{"email@example@example.com", false},
		{".email@example.com", false},
		{"email.@example.com", false},
		{"email..email@example.com", false},
		{"email@example.com (Joe Smith)", false},
		{"email@example", false},
		{"email@111.222.333.44444", false},
		{"email@example..com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v; want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "Secure123!", false},
		{"Short password", "Sec12!", true},
		{"No lowercase", "SECURE123!", true},
		{"No uppercase", "secure123!", true},
		{"No digits", "SecurePass!", true},
		{"No special characters", "SecurePass123", true},
		{"Empty password", "", true},
		{"Unicode support valid", "Sécurisé123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v; wantErr = %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"\n\thello\r\n", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := TrimSpace(tt.input); got != tt.expected {
				t.Errorf("TrimSpace(%q) = %q; expected %q", tt.input, got, tt.expected)
			}
		})
	}
}


