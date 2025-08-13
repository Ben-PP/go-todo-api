package todo

import (
	"errors"
	"fmt"
	"runtime"
	"slices"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (controller *TodoController) UpdateTodo(ctx *gin.Context) {
	payload := &schemas.UpdateTodo{}

	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	} else if payload.Title == nil &&
		payload.Description == nil &&
		payload.CompleteBefore == nil &&
		payload.Completed == nil {
		ctx.JSON(200, gin.H{"status": "not-modified"})
		return
	}

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

	args := &db.GetTodoByIdWithListIdParams{
		ID:     todoID,
		ListID: listID,
	}
	oldTodo, err := controller.db.GetTodoByIdWithListId(ctx, *args)
	if err != nil {
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

	title := oldTodo.Title
	description := oldTodo.Description.String
	completeBefore := &oldTodo.CompleteBefore.Time
	completeBeforeIsValid := oldTodo.CompleteBefore.Valid
	completed := oldTodo.Completed
	if payload.Title != nil {
		title = *payload.Title
	}
	if payload.Description != nil {
		description = *payload.Description
	}
	if payload.CompleteBefore != nil {
		completeBefore = payload.CompleteBefore
		completeBeforeIsValid = true
	}
	if payload.Completed != nil {
		completed = *payload.Completed
	}

	updateArgs := &db.UpdateTodoParams{
		ID:             todoID,
		Title:          title,
		Description:    pgtype.Text{String: description, Valid: true},
		CompleteBefore: pgtype.Timestamp{Time: *completeBefore, Valid: completeBeforeIsValid},
		Completed:      completed,
	}
	newTodo, err := controller.db.UpdateTodo(ctx, *updateArgs)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to update todo",
			file,
			line,
			err,
			ctx,
		)
		return
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventUpdate,
		reqUser,
		&newTodo,
		&oldTodo,
		logging.ObjectEventSubTodo,
	)
	ctx.JSON(200, gin.H{"status": "ok", "todo": newTodo})
}
