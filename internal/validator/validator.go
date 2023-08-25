package validator

import (
	"regexp"
	"strings"
)

// Regex for sanity checking the format of email adresses.
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// For every field check, first start with New() to create a new FieldErrors map
// instance and then use an appropriate method.

// Validator type which contains a map of validations errors.
type Validator struct {
	FieldErrors map[string]string
}

// Create new empty FieldErrors instance with empty FieldErrors map.
func New() *Validator {
	return &Validator{FieldErrors: make(map[string]string)}
}

// Quick check if there's no errors. Returns true if FieldErrors map doesn't
// contain any entries.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

// Adds new error to the map (so long as no entrey already exists for the given
// key)
func (v *Validator) AddFieldError(key, message string) {
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// CheckField() add a message to FieldErrors map only if a validation !ok.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// Check if the field ain't blank.
// If blank then return false and error passed to CheckField().
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// Matches() return true if a string value matches a specific regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
