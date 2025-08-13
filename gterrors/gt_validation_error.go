package gterrors

import "fmt"

type GtValidationError struct {
	Value  string
	Detail string
}

func (e *GtValidationError) Error() string {
	return fmt.Sprintf("validation error: value of %v is invalid. %v", e.Value, e.Detail)
}

func NewGtValueError(value, detail string) *GtValidationError {
	return &GtValidationError{
		Value:  value,
		Detail: detail,
	}
}
