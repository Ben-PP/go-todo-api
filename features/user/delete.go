package user

import (
	"fmt"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/mycontext"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// Controller for deleting users
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
		"deleted",
		userIDToDelete,
		logging.ObjectEventSubUser,
	)
	ctx.JSON(http.StatusNoContent, gin.H{})
}
