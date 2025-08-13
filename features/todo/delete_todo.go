package todo

import (
	"errors"
	"fmt"
	"runtime"
	"slices"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (controller *TodoController) DeleteTodo(ctx *gin.Context) {
	requesterId, requesterUsername, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	listID := ctx.Param("listID")
	todoID := ctx.Param("todoID")
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

	listIds, err := controller.db.GetListIdsAccessible(ctx, reqUser.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to get list accessible by user",
			file,
			line,
			err,
			ctx,
		)
		return
	}
	if !slices.Contains(listIds, listID) {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			fmt.Sprintf("list: %v, todo: %v", listID, todoID),
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	args := &db.DeleteTodoByIdWithListIdParams{
		ID:     todoID,
		ListID: listID,
	}

	if err := controller.db.DeleteTodoByIdWithListId(ctx, *args); err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to delete todo",
			file,
			line,
			err,
			ctx,
		)
		return
	} else {
		logging.LogObjectEvent(
			ctx.FullPath(),
			ctx.ClientIP(),
			logging.ObjectEventDelete,
			reqUser,
			"deleted",
			todoID,
			logging.ObjectEventSubTodo,
		)
	}

	ctx.JSON(204, gin.H{})
}
