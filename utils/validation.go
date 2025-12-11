// Package utils validator for register/login/change user
package utils

import (
	"regexp"

	validationError "backend/internal/errors"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

// Regex declared it so it doesn't get called everytime
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[_-]?[a-zA-Z0-9]+)*$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

type Validator struct{}

// Username function to check length and common regex compatibilty
func (Validator) Username(username string) error {
	if len(username) < 3 {
		return validationError.NewValidatorError("username", "too_short")
	}
	if len(username) > 25 {
		return validationError.NewValidatorError("username", "too_long")
	}
	if !usernameRegex.MatchString(username) {
		return validationError.NewValidatorError("username", "invalid_format")
	}
	return nil
}

// Password function checks if password has enough entropy meaning if it's strong enough
// 60 bits entropy means at least 3 hours to brute force the password
// if the attack has 100,000,000,000,000 guesses per second
// Or 3.7 millions years if he has 10,000
// https://blog.braincoke.fr/security/password-management/
func (Validator) Password(password string) error {
	const minEntropyBits = 60
	err := passwordvalidator.Validate(password, minEntropyBits)
	if err != nil {
		return validationError.NewValidatorError("password", "weak_password")
	}
	return nil
}

// Email function checks if email is a valid email format using a regex expression
func (Validator) Email(email string) error {
	if !emailRegex.MatchString(email) {
		return validationError.NewValidatorError("email", "invalid_format")
	}
	return nil
}
