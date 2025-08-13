package auth

import (
	"errors"
	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/schemas"
	"go-todo/util/mycontext"
	"go-todo/util/passwd"
	"go-todo/util/validate"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (controller *AuthController) UpdatePassword(ctx *gin.Context) {
	userID, _, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	var payload *schemas.UpdatePassword
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}

	isPasswdValid, err := validate.Password(payload.NewPassword)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("error validating password", file, line, err, ctx)
		return
	} else if !isPasswdValid {
		ctx.Error(gterrors.ErrPasswordUnsatisfied).SetType(gin.ErrorTypePublic)
		return
	}

	// Should only fail if something is wrong in the server
	user, err := controller.db.GetUserById(ctx, userID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get user from db", file, line, err, ctx)
		return
	}

	if !passwd.Compare(payload.OldPassword, user.PasswordHash) {
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonInvalidCredentials,
				errors.New("provided credentials are incorrect"),
			),
		).SetType(gin.ErrorTypePublic)
		return
	} else if passwd.Compare(payload.NewPassword, user.PasswordHash) {
		ctx.Error(gterrors.ErrPasswordSame).SetType(gin.ErrorTypePublic)
		return
	}

	newPasswordHash, err := passwd.Hash(payload.NewPassword)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to hash new password", file, line, err, ctx)
		return
	}

	args := &db.UpdateUserPasswordParams{
		PasswordHash: newPasswordHash,
		ID:           user.ID,
	}

	if err := controller.db.UpdateUserPassword(ctx, *args); err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to update password to db", file, line, err, ctx)
		return
	}

	refreshToken, refreshClaims, accessToken, _, err := generateTokens(
		"",
		user,
	)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		failedToGenerateJwtError(err, file, line, ctx)
		return
	}

	refreshArgs := &db.CreateJwtTokenParams{
		Jti:       refreshClaims.ID,
		UserID:    refreshClaims.Subject,
		Family:    refreshClaims.Family,
		CreatedAt: pgtype.Timestamp{Time: refreshClaims.IssuedAt.Time, Valid: true},
		ExpiresAt: pgtype.Timestamp{Time: refreshClaims.ExpiresAt.Time, Valid: true},
	}

	if err := controller.db.CreateJwtToken(ctx, *refreshArgs); err != nil {
		_, file, line, _ := runtime.Caller(0)
		failedToSaveJwtToDbError(err, file, line, ctx)
		return
	}

	deleteArgs := &db.DeleteJwtTokenByUserIdExcludeFamilyParams{
		UserID: userID,
		Family: refreshClaims.Family,
	}

	if err := controller.db.DeleteJwtTokenByUserIdExcludeFamily(ctx, *deleteArgs); err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to remove old refresh jwts", file, line, err, ctx)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":        "ok",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
