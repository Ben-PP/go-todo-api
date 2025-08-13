package auth

import (
	"errors"
	"fmt"
	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/jwt"
	"go-todo/util/mycontext"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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
	ctx.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
