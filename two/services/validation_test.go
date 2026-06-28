package services

import "testing"

func TestIsValidEmail_Valid(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"john.doe@company.org",
		"info@sub.domain.co.uk",
	}

	for _, email := range validEmails {
		if !IsValidEmail(email) {
			t.Errorf("Expected email %q to be valid, but was invalid", email)
		}
	}
}

func TestIsValidEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"plainaddress",
		"#@%^%#$@#$@#.com",
		"@example.com",
		"Joe Smith <email@example.com>",
		"email.example.com",
		"email@example@example.com",
		"email@example.c",
	}

	for _, email := range invalidEmails {
		if IsValidEmail(email) {
			t.Errorf("Expected email %q to be invalid, but was valid", email)
		}
	}
}

func TestValidatePassword_Valid(t *testing.T) {
	err := ValidatePassword("StrongPass123!")
	if err != nil {
		t.Errorf("Expected password to be valid, got error: %v", err)
	}
}

func TestValidatePassword_Invalid(t *testing.T) {
	tests := []struct {
		password string
		expected string
	}{
		{"Short1!", "password must be at least 8 characters long"},
		{"lowercaseonly1!", "password must contain at least one uppercase letter"},
		{"UPPERCASEONLY1!", "password must contain at least one lowercase letter"},
		{"NoSpecialChar123", "password must contain at least one special character"},
		{"NoNumberCheck!", "password must contain at least one number"},
	}

	for _, tc := range tests {
		err := ValidatePassword(tc.password)
		if err == nil {
			t.Errorf("Expected password %q to be invalid, but got no error", tc.password)
		} else if err.Error() != tc.expected {
			t.Errorf("Expected error %q for password %q, got %q", tc.expected, tc.password, err.Error())
		}
	}
}

func TestTrimSpace(t *testing.T) {
	expected := "hello"
	actual := TrimSpace("  hello  ")
	if actual != expected {
		t.Errorf("Expected trimmed string %q, got %q", expected, actual)
	}
}

func TestFailingExampleTwo_Invalid(t *testing.T) {
	// Let's add a failing test in project two as well to verify both jobs
	// are captured correctly in the failure summary.
	t.Log("Deliberate failure in project two services test")
	t.Fail()
}
