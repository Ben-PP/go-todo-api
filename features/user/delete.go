package user

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

// DeleteUser handles user deletion requests.
//
//	@Summary		Delete user
//	@Description	Deletes a user account. Users can delete their own accounts; admins can delete any account.
//	@Tags			User
//	@Security		Bearer
//	@Produce		json
//	@Param			id	path		string	true	"ID of the user to delete"
//	@Success		204	{object}	nil	"Indicates successful deletion with no content."
//	@Failure		400	{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401	{object}	schemas.ErrorResponse	"Unauthorized due to invalid credentials."
//	@Failure		403	{object}	schemas.ErrorResponse	"Forbidden action."
//	@Failure		500	{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/user/{id} [delete]
func (controller *UserController) DeleteUser(ctx *gin.Context) {
	tokenUserId, tokenUserName, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	reqUser, err := controller.db.GetUserById(ctx, tokenUserId)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			tokenUserName,
			ctx.ClientIP(),
		)
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonJwtUserNotFound,
				fmt.Errorf("could not get user from db: %w", err),
			),
		).SetType(gterrors.GetGinErrorType())
		return
	}

	userIDToDelete := ctx.Param("id")
	if userIDToDelete != reqUser.ID && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreMedium,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			userIDToDelete,
			reqUser.Username,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gterrors.GetGinErrorType())
		return
	}

	rows, err := controller.db.DeleteUser(ctx, userIDToDelete)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("could not delete user", file, line, err, ctx)
		return
	}
	if rows == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "user-not-removed"})
		return
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventDelete,
		&reqUser,
		userIDToDelete,
		reflect.TypeOf(db.User{}).String(),
	)
	ctx.JSON(http.StatusNoContent, gin.H{})
}
