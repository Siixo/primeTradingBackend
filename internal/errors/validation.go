// Package errors to format the errors return statements
package errors

import "strings"

type ValidationError struct {
	Field string
	Code  string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Code
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var listErrors strings.Builder
	for _, e := range v {
		listErrors.WriteString(e.Error())
		listErrors.WriteString(": ")
	}
	return listErrors.String()
}

func NewValidatorError(field, code string) ValidationError {
	return ValidationError{
		Field: field,
		Code:  code,
	}
}
