package auth

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/jwt"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Refresh handles token refresh requests.
//
//	@Summary		Refresh tokens
//	@Description	Refreshes JWT access and refresh tokens using a valid refresh token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh	body		schemas.Refresh			true	"Refresh token"
//	@Success		200		{object}	schemas.LoginResponse	"Returns new access and refresh tokens."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to invalid or expired token."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/auth/refresh [post]
func (controller *AuthController) Refresh(ctx *gin.Context) {
	var payload *schemas.Refresh
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}

	refreshToken := payload.RefreshToken

	decodedRefreshToken, err := jwt.DecodeRefreshToken(refreshToken)
	if err != nil {
		logTokenEventUse(false, decodedRefreshToken, ctx)
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

			ctx.Error(gterrors.NewGtAuthError(reason, err)).SetType(gterrors.GetGinErrorType())
			return
		}
		// Should never get to here
		ctx.Error(gterrors.ErrShouldNotHappen)
		return
	}

	dbToken, err := controller.db.GetJwtTokenByJti(ctx, decodedRefreshToken.ID)
	if err != nil {
		logTokenEventUse(false, decodedRefreshToken, ctx)
		logging.LogSecurityEvent(
			logging.SecurityScoreMedium,
			logging.SecurityEventJwtUnknown,
			ctx.FullPath(),
			decodedRefreshToken.ID,
			ctx.ClientIP(),
		)
		ginType := gterrors.GetGinErrorType()
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(
				gterrors.NewGtAuthError(gterrors.GtAuthErrorReasonTokenInvalid, err),
			).SetType(ginType)
			return
		}

		ctx.Error(gterrors.NewGtAuthError(
			gterrors.GtAuthErrorReasonInternalError,
			fmt.Errorf("failed to get token from db: %w", err),
		)).SetType(ginType)
		return
	}

	if dbToken.IsUsed {
		logTokenEventUse(false, decodedRefreshToken, ctx)
		logging.LogSecurityEvent(
			logging.SecurityScoreCritical,
			logging.SecurityEventJwtReuse,
			ctx.FullPath(),
			decodedRefreshToken.ID,
			ctx.ClientIP(),
		)
		ginType := gterrors.GetGinErrorType()
		if rows, err := controller.db.DeleteJwtTokenByFamily(ctx, dbToken.Family); err != nil || rows == 0 {
			_, file, line, _ := runtime.Caller(0)
			errIfNil := fmt.Errorf("failed to delete jwt family: %w", err)
			if err == nil {
				errIfNil = fmt.Errorf("failed to delete jwt family: %v", dbToken.Family)
			}
			mycontext.CtxAddGtInternalError("", file, line, errIfNil, ctx)
			return
		}
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonTokenReuse,
				gterrors.ErrJwtRefreshReuse,
			),
		).SetType(ginType)
		return
	}

	// This should always succeed if db works correctly as dbToken has to have
	// userID. Errors are system failures.
	user, err := controller.db.GetUserById(ctx, dbToken.UserID)
	if err != nil {
		logTokenEventUse(false, decodedRefreshToken, ctx)
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get user from db", file, line, err, ctx)
		return
	}

	logSessionRefresh := func(success bool) {
		logging.LogSessionEvent(
			success,
			ctx.FullPath(),
			user.Username,
			logging.SessionEventTypeRefresh,
			ctx.ClientIP(),
		)
	}

	refreshToken, refreshClaims, accessToken, accessClaims, err := generateTokens(
		decodedRefreshToken.Family,
		user,
	)
	if err != nil {
		logSessionRefresh(false)
		logTokenEventUse(false, decodedRefreshToken, ctx)
		_, file, line, _ := runtime.Caller(0)
		failedToGenerateJwtError(err, file, line, ctx)
		return
	}

	// Mark the token as used.
	if err := controller.db.UseJwtToken(ctx, dbToken.Jti); err != nil {
		logTokenEventUse(false, decodedRefreshToken, ctx)
		logSessionRefresh(false)
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("", file, line, err, ctx)
		return
	}

	args := &db.CreateJwtTokenParams{
		Jti:       refreshClaims.ID,
		UserID:    refreshClaims.Subject,
		Family:    refreshClaims.Family,
		CreatedAt: pgtype.Timestamp{Time: refreshClaims.IssuedAt.Time, Valid: true},
		ExpiresAt: pgtype.Timestamp{Time: refreshClaims.ExpiresAt.Time, Valid: true},
	}

	if err := controller.db.CreateJwtToken(ctx, *args); err != nil {
		logTokenEventUse(false, decodedRefreshToken, ctx)
		logSessionRefresh(false)
		_, file, line, _ := runtime.Caller(0)
		failedToSaveJwtToDbError(err, file, line, ctx)
		return
	}

	logSessionRefresh(true)
	logTokenCreations([]*jwt.GtClaims{refreshClaims, accessClaims}, ctx)
	logTokenEventUse(true, decodedRefreshToken, ctx)
	response := &schemas.LoginResponse{
		Status:       "ok",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	ctx.JSON(http.StatusOK, response)
}
