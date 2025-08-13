package todo

import (
	"fmt"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

func (controller *TodoController) DeleteList(ctx *gin.Context) {
	requesterId, requesterUsername, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	listID := ctx.Param("listID")
	reqUser, err := database.GetUserById(controller.db, requesterId, ctx)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			requesterUsername,
			ctx.ClientIP(),
		)
		return
	}

	listDeleted, err := controller.db.GetList(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to get list",
			file,
			line,
			err,
			ctx,
		)
		return
	}

	if listDeleted.UserID != reqUser.ID && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			fmt.Sprintf("listID: %v", listID),
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	rows, err := controller.db.DeleteList(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to delete list",
			file,
			line,
			err,
			ctx,
		)
		return
	}

	if rows != 0 {
		logging.LogObjectEvent(
			ctx.FullPath(),
			ctx.ClientIP(),
			logging.ObjectEventDelete,
			reqUser,
			"deleted",
			listDeleted.ID,
			logging.ObjectEventSubList,
		)
	}
	ctx.JSON(204, gin.H{})
}
