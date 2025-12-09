package todo

import (
	"errors"
	"reflect"
	"runtime"
	"slices"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// ReadListWithTodos handles the retrieval of a todo list along with its associated todos.
//
//	@Summary		Get todo list with todos
//	@Description	Retrieves a todo list by its ID along with all associated todo items. Access is restricted to users with permissions for the list.
//	@Tags			Lists
//	@Security		Bearer
//	@Produce		json
//	@Param			listID	path		string	true	"ID of the todo list to retrieve"
//	@Success		200		{object}	schemas.ListWithTodos	"Returns the todo list details along with its todos."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for users without access to the list."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID} [get]
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

	response := schemas.ListWithTodos{
		List:  list,
		Todos: todos,
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventRead,
		reqUser,
		list.ID,
		reflect.TypeOf(list).String(),
	)
	ctx.JSON(200, response)
}
