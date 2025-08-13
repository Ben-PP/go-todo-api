package auth

import (
	"errors"
	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/jwt"
	"go-todo/util/mycontext"
	"go-todo/util/passwd"
	"go-todo/util/validate"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (controller *AuthController) Login(ctx *gin.Context) {
	var payload *schemas.Login
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}

	username := payload.Username
	password := payload.Password
	ok, err := validate.Username(username)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("", file, line, err, ctx)
		return
	} else if !ok {
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonUsernameInvalid,
				errors.New("username validation failed"),
			),
		).SetType(gterrors.GetGinErrorType())
		return
	}

	user, err := controller.db.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logging.LogSecurityEvent(
				logging.SecurityScoreLow,
				logging.SecurityEventLoginToUnknownUsername,
				ctx.FullPath(),
				username,
				ctx.ClientIP(),
			)

			ctx.Error(
				gterrors.NewGtAuthError(
					gterrors.GtAuthErrorReasonInvalidCredentials,
					err,
				),
			).SetType(gterrors.GetGinErrorType())
			return
		}

		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get username from db", file, line, err, ctx)
		return
	}

	if pwdCorrect := passwd.Compare(password, user.PasswordHash); !pwdCorrect {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventFailedLogin,
			ctx.FullPath(),
			username,
			ctx.ClientIP(),
		)
		logging.LogSessionEvent(
			false,
			ctx.FullPath(),
			user.Username,
			logging.SessionEventTypeLogin,
			ctx.ClientIP(),
		)

		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonInvalidCredentials,
				errors.New("password verification failed"),
			),
		).SetType(gterrors.GetGinErrorType())
		return
	}

	refreshToken, refreshClaims, accessToken, accessClaims, err := generateTokens(
		"",
		user,
	)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		failedToGenerateJwtError(err, file, line, ctx)
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
		_, file, line, _ := runtime.Caller(0)
		failedToSaveJwtToDbError(err, file, line, ctx)
		return
	}

	logging.LogSessionEvent(
		true,
		ctx.FullPath(),
		user.Username,
		logging.SessionEventTypeLogin,
		ctx.ClientIP(),
	)
	logTokenCreations([]*jwt.GtClaims{refreshClaims, accessClaims}, ctx)
	ctx.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
