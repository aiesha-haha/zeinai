package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// safeNameRegex allows Unicode letters, spaces, hyphens, and apostrophes.
// It explicitly rejects any HTML special characters (<, >, &, ", ').
var safeNameRegex = regexp.MustCompile(`^[\p{L}\s'\-]+$`)

// NewValidator returns a *validator.Validate instance with all custom
// rules registered. Use this instead of validator.New() everywhere.
func NewValidator() *validator.Validate {
	v := validator.New()

	// safe_name: only letters (any Unicode script), spaces, hyphens, apostrophes.
	// Blocks HTML/JS injection like <img src=x onerror=alert(1)>.
	if err := v.RegisterValidation("safe_name", validateSafeName); err != nil {
		panic("validation: failed to register safe_name rule: " + err.Error())
	}

	return v
}

// validateSafeName is the custom validator function for the safe_name tag.
func validateSafeName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // let `required` handle empty checks
	}
	return safeNameRegex.MatchString(value)
}
