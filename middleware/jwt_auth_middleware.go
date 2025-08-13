package middleware

import (
	"errors"
	"fmt"
	"go-todo/gterrors"
	"go-todo/logging"
	jwtUtil "go-todo/util/jwt"
	"runtime"

	"github.com/gin-gonic/gin"
)

// Tries to extract the JWT from Authorization header. Returns an error status
// to the client if it fails.
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := jwtUtil.DecodeTokenFromHeader(c)
		if err != nil {
			ginType := gterrors.GetGinErrorType()
			var jwtErr *jwtUtil.JwtDecodeError
			if errors.As(err, &jwtErr) {
				if jwtErr.Claims != nil {
					logging.LogTokenEvent(false, c.FullPath(), logging.TokenEventTypeAccess, c.RemoteIP(), jwtErr.Claims)
				}
				reason := gterrors.GtAuthErrorReasonInternalError

				switch jwtErr.Reason {
				case jwtUtil.JwtErrorReasonExpired:
					reason = gterrors.GtAuthErrorReasonExpired
				case jwtUtil.JwtErrorReasonInvalidSignature:
					reason = gterrors.GtAuthErrorReasonInvalidSignature
				case jwtUtil.JwtErrorReasonUnhandled:
					reason = gterrors.GtAuthErrorReasonInternalError
				default:
					reason = gterrors.GtAuthErrorReasonInternalError
				}

				c.Error(gterrors.NewGtAuthError(reason, jwtErr.Err)).SetType(ginType)
				c.Abort()
				return
			} else {
				// Should never happen
				_, file, line, _ := runtime.Caller(0)
				c.Error(
					gterrors.NewGtInternalError(
						errors.Join(gterrors.ErrShouldNotHappen, err),
						fmt.Sprintf("%v: %d", file, line),
						500,
					),
				)
			}
			c.Abort()
			return
		}

		logging.LogTokenEvent(true, c.FullPath(), logging.TokenEventTypeAccess, c.RemoteIP(), token)

		c.Set("x-token-username", token.Username)
		c.Set("x-token-user-id", token.Subject)
		c.Set("x-token-is-admin", token.IsAdmin)

		c.Next()
	}
}
