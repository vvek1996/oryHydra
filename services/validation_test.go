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
