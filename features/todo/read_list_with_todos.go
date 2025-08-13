package todo

import (
	"errors"
	"runtime"
	"slices"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (controller *TodoController) ReadListWithTodos(ctx *gin.Context) {
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
	allowedIds, err := controller.db.GetListIdsAccessible(ctx, reqUser.ID)
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
	if !slices.Contains(allowedIds, listID) && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			listID,
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	list, err := controller.db.GetList(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get list", file, line, err, ctx)
		return
	}

	todos, err := controller.db.GetTodosByList(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get todos", file, line, err, ctx)
		return
	}

	response := map[string]any{
		"id":          list.ID,
		"user_id":     list.UserID,
		"title":       list.Title,
		"description": list.Description,
		"created_at":  list.CreatedAt,
		"updated_at":  list.UpdatedAt,
		"todos":       todos,
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventRead,
		reqUser,
		&list,
		nil,
		logging.ObjectEventSubList,
	)
	ctx.JSON(200, gin.H{"status": "ok", "list": response})
}
