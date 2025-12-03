package auth

import (
	"errors"
	"net/http"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/schemas"
	"go-todo/util/mycontext"
	"go-todo/util/passwd"
	"go-todo/util/validate"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpdatePassword handles user password update requests.
//
//	@Summary		Update user password
//	@Description	Updates the password for an authenticated user.
//	@Security		Bearer
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			updatePassword	body		schemas.UpdatePassword	true	"Old and new passwords"
//	@Success		200				{object}	schemas.LoginResponse	"Returns new access and refresh tokens."
//	@Failure		400				{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401				{object}	schemas.ErrorResponse	"Unauthorized due to invalid token or invalid password."
//	@Failure		500				{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/auth/update-password [post]
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

	response := &schemas.LoginResponse{
		Status:       "ok",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	ctx.JSON(http.StatusOK, response)
}
