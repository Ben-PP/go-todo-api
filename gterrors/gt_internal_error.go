package gterrors

import "fmt"

type GtInternalError struct {
	Err            error
	File           string
	ResponseStatus int
}

func (e *GtInternalError) Error() string {
	return fmt.Sprintf("internal error: %v", e.Err.Error())
}

func NewGtInternalError(err error, file string, responseStatus int) *GtInternalError {
	return &GtInternalError{
		Err:            err,
		File:           file,
		ResponseStatus: responseStatus,
	}
}
