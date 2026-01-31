package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrRequiredField    = errors.New("field is required")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrInvalidRating    = errors.New("rating must be between 1 and 5")
	ErrInvalidObjectID  = errors.New("invalid ID format")
)

// emailRegex is a simple regex for email validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateRequired checks if a string field is not empty
func ValidateRequired(field, fieldName string) error {
	if strings.TrimSpace(field) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword validates password requirements
func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	if len(password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

// ValidateRating validates that rating is between 1 and 5
func ValidateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return ErrInvalidRating
	}
	return nil
}

// ValidateMinLength validates minimum length of a string
func ValidateMinLength(field, fieldName string, minLen int) error {
	if len(strings.TrimSpace(field)) < minLen {
		return errors.New(fieldName + " must be at least " + string(rune(minLen+'0')) + " characters")
	}
	return nil
}

// ValidationErrors holds multiple validation errors
type ValidationErrors struct {
	Errors []string `json:"errors"`
}

// Add adds an error to the validation errors
func (v *ValidationErrors) Add(err error) {
	if err != nil {
		v.Errors = append(v.Errors, err.Error())
	}
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error returns a combined error message
func (v *ValidationErrors) Error() string {
	return strings.Join(v.Errors, "; ")
}
