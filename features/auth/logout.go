package auth

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/jwt"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

// Logout handles user logout requests.
//
//	@Summary		Logout user
//	@Description	Logs out a user by invalidating their refresh token.
//	@Security		Bearer
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			logout	body	schemas.Refresh	true	"Refresh token to invalidate"
//	@Success		204		"No Content - Successfully logged out."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to invalid or expired token."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/auth/logout [post]
func (controller *AuthController) Logout(ctx *gin.Context) {
	// TODO Add access jwt to redis blacklist
	var payload *schemas.Refresh
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}

	refreshToken := payload.RefreshToken

	claims, err := jwt.DecodeRefreshToken(refreshToken)
	if err != nil {
		logTokenEventUse(false, claims, ctx)
		var jwtErr *jwt.JwtDecodeError
		if errors.As(err, &jwtErr) {
			reason := gterrors.GtAuthErrorReasonInternalError
			switch jwtErr.Reason {
			case jwt.JwtErrorReasonExpired:
				reason = gterrors.GtAuthErrorReasonExpired
			case jwt.JwtErrorReasonInvalidSignature:
				reason = gterrors.GtAuthErrorReasonInvalidSignature
			case jwt.JwtErrorReasonTokenMalformed:
				reason = gterrors.GtAuthErrorReasonTokenInvalid
			case jwt.JwtErrorReasonUnhandled:
				reason = gterrors.GtAuthErrorReasonInternalError
			}

			ctx.Error(
				gterrors.NewGtAuthError(
					reason,
					errors.Join(gterrors.ErrGtLogoutFailure, err),
				),
			).SetType(gterrors.GetGinErrorType())
			return
		}
		// Should never get to here
		ctx.Error(gterrors.ErrShouldNotHappen)
		return
	}

	userID, violatorName, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	} else if userID != claims.Subject {
		logging.LogSecurityEvent(
			logging.SecurityScoreHigh,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			claims.Username,
			violatorName,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gterrors.GetGinErrorType())
		return
	}

	if rows, err := controller.db.DeleteJwtTokenByFamily(ctx, claims.Family); err != nil || rows == 0 {
		_, file, line, _ := runtime.Caller(0)
		errIfNil := fmt.Errorf("failed to delete jwt family: %w", err)
		if err == nil {
			errIfNil = fmt.Errorf("failed to delete jwt family: %v", claims.Family)
		}
		mycontext.CtxAddGtInternalError("", file, line, errIfNil, ctx)
		return
	}

	logging.LogSessionEvent(
		true,
		ctx.FullPath(),
		claims.Username,
		logging.SessionEventTypeLogout,
		ctx.ClientIP(),
	)
	ctx.JSON(http.StatusNoContent, gin.H{})
}
