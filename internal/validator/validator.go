// internal/validator/validator.go
package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// EmailRX is a compiled regular expression for basic email format validation.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator struct holds a map of validation errors.
// The key is the field name ("email") and the value is the error message.
type Validator struct {
	Errors map[string]string
}

// NewValidator creates and returns a new Validator instance with an initialized Errors map.
func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// ValidData returns true if the Errors map is empty (no validation errors), false otherwise
func (v *Validator) ValidData() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the Errors map for a specific field,
// but only if an error for that field doesn't already exist.
// This prevents multiple errors from being added for the same field during a single validation pass.
func (v *Validator) AddError(field string, message string) {
	_, exists := v.Errors[field]
	if !exists {
		v.Errors[field] = message
	}
}

// Check is a helper method. If 'ok' is false, it calls AddError with the given field and message.
func (v *Validator) Check(ok bool, field string, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// --- Standalone Validation Helper Functions ---

// NotBlank returns true if a string is not empty after trimming leading/trailing whitespace.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MinLength returns true if a string contains at least 'n' runes (characters).
// It uses utf8.RuneCountInString to correctly handle multi-byte characters.
func MinLength(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxLength returns true if a string contains at most 'n' runes.
// It uses utf8.RuneCountInString for correct character counting.
func MaxLength(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// IsValidEmail returns true if a string matches the EmailRX regular expression pattern.
func IsValidEmail(email string) bool {
	return EmailRX.MatchString(email)
}