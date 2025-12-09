package todo

import (
	"fmt"
	"reflect"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

// DeleteList handles the deletion of a todo list.
//
//	@Summary		Delete todo list
//	@Description	Deletes a todo list by its ID. Only the owner or an admin can delete the list.
//	@Tags			Lists
//	@Security		Bearer
//	@Produce		json
//	@Param			listID	path		string	true	"ID of the todo list to delete"
//	@Success		204		"No content, indicating successful deletion."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for non-owners/non-admins."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID} [delete]
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
			listDeleted.ID,
			reflect.TypeOf(listDeleted).String(),
		)
	}
	ctx.JSON(204, gin.H{})
}
