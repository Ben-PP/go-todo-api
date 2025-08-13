package gterrors

import (
	"errors"
	"fmt"
)

var ErrGtLogoutFailure = errors.New("failed to logout")

type GtAuthErrorReason int

const (
	GtAuthErrorReasonExpired GtAuthErrorReason = iota
	GtAuthErrorReasonInternalError
	GtAuthErrorReasonInvalidCredentials
	GtAuthErrorReasonInvalidSignature
	GtAuthErrorReasonTokenInvalid
	GtAuthErrorReasonJwtUserNotFound
	GtAuthErrorReasonTokenReuse
	GtAuthErrorReasonUsernameInvalid
)

func (t GtAuthErrorReason) String() string {
	switch t {
	case GtAuthErrorReasonExpired:
		return "token-expired"
	case GtAuthErrorReasonJwtUserNotFound:
		return "jwt-user-not-found"
	case GtAuthErrorReasonTokenInvalid:
		return "token-invalid"
	case GtAuthErrorReasonInvalidCredentials:
		return "invalid-credentials"
	case GtAuthErrorReasonInvalidSignature:
		return "token-invalid-signature"
	case GtAuthErrorReasonInternalError:
		return "internal-server-error"
	case GtAuthErrorReasonTokenReuse:
		return "jwt-token-reuse"
	case GtAuthErrorReasonUsernameInvalid:
		return "username-invalid"
	}
	return "unknown"
}

type GtAuthError struct {
	Reason GtAuthErrorReason
	Err    error
}

func (e *GtAuthError) Error() string {
	return fmt.Sprintf("failed to validate jwt: %v", e.Reason.String())
}

func NewGtAuthError(reason GtAuthErrorReason, err error) *GtAuthError {
	return &GtAuthError{
		Reason: reason,
		Err:    err,
	}
}
