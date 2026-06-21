package services

import (
	"net/mail"
	"strings"
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
