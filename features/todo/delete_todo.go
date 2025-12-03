package todo

import (
	"errors"
	"fmt"
	"reflect"
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

// DeleteTodo handles the deletion of a todo item.
//
//	@Summary		Delete todo item
//	@Description	Deletes a todo item by its ID within a specified list. Only users with access to the list can delete the todo item.
//	@Tags			Todos
//	@Security		Bearer
//	@Produce		json
//	@Param			listID	path		string	true	"ID of the todo list containing the todo item"
//	@Param			todoID	path		string	true	"ID of the todo item to delete"
//	@Success		204		{object}	nil		"No content, indicating successful deletion."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for users without access to the list."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo item or list not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID}/todo/{todoID} [delete]
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
			todoID,
			reflect.TypeOf(db.Todo{}).String(),
		)
	}

	ctx.JSON(204, gin.H{})
}
