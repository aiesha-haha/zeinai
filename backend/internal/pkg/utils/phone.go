package utils

import "strings"

// NormalizePhone removes all non-digit characters for consistent storage and lookup
func NormalizePhone(phone string) string {
	var res strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			res.WriteRune(r)
		}
	}
	return res.String()
}

// IsEmail checks if identifier is an email
func IsEmail(identifier string) bool {
	return strings.Contains(identifier, "@")
}
