package services

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

// IsValidEmail checks if an email string has a valid format according to RFC 5322.
// It also ensures the address is not empty and contains a domain name with a valid alphabetical TLD.
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// mail.ParseAddress can accept "Name <email@domain.com>", so we check if the parsed Address 
	// matches the input exactly.
	if addr.Address != email {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	
	domain := parts[1]
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return false
	}

	// Validate each domain label is not empty
	for _, label := range domainParts {
		if label == "" {
			return false
		}
	}

	// The TLD (last part of domain) must contain only letters and be at least 2 chars
	tld := domainParts[len(domainParts)-1]
	if len(tld) < 2 {
		return false
	}
	for _, ch := range tld {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}

	return true
}

// ValidatePassword checks if a password meets standard security requirements:
// - at least 8 characters long
// - at least one lowercase letter
// - at least one uppercase letter
// - at least one digit
// - at least one special character
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var (
		hasLower   bool
		hasUpper   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch): // unicode.IsDigit is more appropriate than unicode.IsNumber to match standard 0-9 digits
			hasNumber = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// TrimSpace removes leading and trailing white spaces from a string.
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}


