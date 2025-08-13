package middleware

import (
	"errors"
	"fmt"
	"go-todo/gterrors"
	"go-todo/logging"
	"runtime"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"
)

type ResponseParams struct {
	Status        int
	StatusMessage string
	Detail        string
}

type StatusMessage int

const (
	StatusMessageForbidden StatusMessage = iota
	StatusMessageInternalServerError
	StatusMessageInvalidCredentials
	StatusMessageMalformedBody
	StatusMessageNotFound
	StatusMessagePasswordUnsatisfied
	StatusMessageUnauthorized
	StatusMessageUniqueViolation
	StatusMessageUsernameUnsatisfied
)

func (t StatusMessage) String() string {
	switch t {
	case StatusMessageForbidden:
		return "forbidden"
	case StatusMessageInternalServerError:
		return "internal-server-error"
	case StatusMessageInvalidCredentials:
		return "invalid-credentials"
	case StatusMessageMalformedBody:
		return "malformed-body"
	case StatusMessageNotFound:
		return "not-found"
	case StatusMessagePasswordUnsatisfied:
		return "password-unsatisfied"
	case StatusMessageUnauthorized:
		return "unauthorized"
	case StatusMessageUniqueViolation:
		return "unique-violation"
	case StatusMessageUsernameUnsatisfied:
		return "username-unsatisfied"
	}
	return "unknown-status-message"
}

// Handles the errors passed to the gin context and responds accordingly.
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last()
		isPublic := err.Type == gin.ErrorTypePublic
		var params *ResponseParams
		var authError *gterrors.GtAuthError
		var internalError *gterrors.GtInternalError
		var validationError *gterrors.GtValidationError
		switch {
		// Malformed requests
		case err.Type == gin.ErrorTypeBind:
			params = &ResponseParams{400, StatusMessageMalformedBody.String(), err.Error()}
		case errors.Is(err, gterrors.ErrPasswordUnsatisfied) || errors.Is(err, gterrors.ErrPasswordSame):
			params = &ResponseParams{400, StatusMessagePasswordUnsatisfied.String(), err.Error()}
		case errors.Is(err, gterrors.ErrUsernameUnsatisfied):
			params = &ResponseParams{400, StatusMessageUsernameUnsatisfied.String(), err.Error()}
		case errors.Is(err, gterrors.ErrForbidden):
			params = &ResponseParams{403, StatusMessageForbidden.String(), err.Error()}
		case errors.Is(err, gterrors.ErrUniqueViolation):
			params = &ResponseParams{409, StatusMessageUniqueViolation.String(), err.Error()}
		case errors.Is(err, gterrors.ErrNotFound):
			params = &ResponseParams{404, StatusMessageNotFound.String(), err.Error()}
		case errors.As(err, &validationError):
			params = &ResponseParams{
				400,
				"invalid-value",
				validationError.Error(),
			}
		// GtAuthError
		case errors.As(err.Err, &authError):
			detail := ""
			status := 401
			var statusMessage string
			if isPublic {
				detail = authError.Reason.String()
			}

			switch authError.Reason {
			case gterrors.GtAuthErrorReasonExpired:
				statusMessage = StatusMessageUnauthorized.String()
			case gterrors.GtAuthErrorReasonJwtUserNotFound:
				statusMessage = StatusMessageUnauthorized.String()
			case gterrors.GtAuthErrorReasonInvalidCredentials:
				statusMessage = StatusMessageInvalidCredentials.String()
			case gterrors.GtAuthErrorReasonInvalidSignature:
				statusMessage = StatusMessageUnauthorized.String()
			case gterrors.GtAuthErrorReasonTokenInvalid:
				statusMessage = StatusMessageUnauthorized.String()
			case gterrors.GtAuthErrorReasonTokenReuse:
				statusMessage = StatusMessageUnauthorized.String()
			case gterrors.GtAuthErrorReasonUsernameInvalid:
				statusMessage = StatusMessageInvalidCredentials.String()
			case gterrors.GtAuthErrorReasonInternalError:
				statusMessage = StatusMessageUnauthorized.String()
				if isPublic {
					detail = authError.Err.Error()
				}
				logging.LogError(authError.Err, "unknown", authError.Error())
			default:
				statusMessage = StatusMessageInternalServerError.String()
				status = 500
				if isPublic {
					detail = authError.Err.Error()
				}
			}

			// If error originates from logout event
			if errors.Is(authError.Err, gterrors.ErrGtLogoutFailure) {
				statusMessage = StatusMessageUnauthorized.String()
				if isPublic {
					detail = authError.Err.Error()
				}
			}

			params = &ResponseParams{
				Status:        status,
				StatusMessage: statusMessage,
				Detail:        detail,
			}
		// GtInternalError
		case errors.As(err, &internalError):
			logging.LogError(internalError.Err, internalError.File, "")
			detail := ""
			if isPublic {
				detail = internalError.Error()
			}
			var statusMessage string
			switch internalError.ResponseStatus {
			case 500:
				statusMessage = StatusMessageInternalServerError.String()
			default:
				statusMessage = StatusMessageInternalServerError.String()
			}
			params = &ResponseParams{
				Status:        internalError.ResponseStatus,
				StatusMessage: statusMessage,
				Detail:        detail,
			}
		// Catch all
		case errors.Is(err, gterrors.ErrShouldNotHappen):
			params = &ResponseParams{
				Status:        500,
				StatusMessage: "should-never-happen",
				Detail:        "Congratulations! You have caused an error that should not be possible to happen :D",
			}
		default:
			detail := ""
			errToShow := fmt.Errorf("unknown error: %w", err)
			if isPublic {
				detail = errToShow.Error()
			}
			params = &ResponseParams{
				Status:        500,
				StatusMessage: StatusMessageInternalServerError.String(),
				Detail:        detail,
			}
			_, file, line, _ := runtime.Caller(0)
			logging.LogError(errToShow, fmt.Sprintf("%v: %d", file, line), "")
		}

		body := gin.H{"status": params.StatusMessage}
		if params.Detail != "" {
			body = gin.H{
				"status": params.StatusMessage,
				"detail": params.Detail,
			}
		}
		c.JSON(params.Status, body)
	}
}
