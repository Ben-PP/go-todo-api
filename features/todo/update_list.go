package todo

import (
	"errors"
	"reflect"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"
	"go-todo/util/validate"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpdateList handles updating a todo list's title and/or description.
//
//	@Summary		Update a todo list
//	@Description	Updates the title and/or description of a specific todo list. Requires ownership of the list or admin privileges.
//	@Tags			Lists
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			listID	path		string					true	"ID of the todo list to update"
//	@Param			update	body		schemas.UpdateList		true	"Updated title and/or description"
//	@Success		200		{object}	db.List					"Returns the updated todo list."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for non-owners or non-admins."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID} [patch]
func (controller *TodoController) UpdateList(ctx *gin.Context) {
	var payload *schemas.UpdateList
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	} else if payload.Description == nil && payload.Title == nil {
		ctx.Error(errors.New("either title or description is required")).SetType(gin.ErrorTypeBind)
		return
	}

	listID := ctx.Param("listID")

	tokenUserId, tokenUserName, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	reqUser, err := database.GetUserById(controller.db, tokenUserId, ctx)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			tokenUserName,
			ctx.ClientIP(),
		)
		return
	}

	oldList, err := controller.db.GetList(ctx, listID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
			return
		}
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("could not get user from db", file, line, err, ctx)
		return
	}

	if oldList.UserID != reqUser.ID && !reqUser.IsAdmin {
		ctx.Error(gterrors.ErrForbidden).SetType(gterrors.GetGinErrorType())
		return
	}

	title := oldList.Title
	description := oldList.Description.String
	if payload.Title != nil {
		title = *payload.Title
	}
	if payload.Description != nil {
		description = *payload.Description
	}
	if !validate.LengthTitle(title) {
		ctx.Error(gterrors.NewGtValueError(title, "title too long"))
		return
	} else if !validate.LengthDescription(description) {
		ctx.Error(gterrors.NewGtValueError(description, "description too long"))
		return
	}

	args := &db.UpdateListParams{
		Title:       title,
		Description: pgtype.Text{String: description, Valid: payload.Description != nil},
		ID:          listID,
	}

	newList, err := controller.db.UpdateList(ctx, *args)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to update list", file, line, err, ctx)
		return
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventUpdate,
		reqUser,
		newList.ID,
		reflect.TypeOf(newList).String(),
	)
	ctx.JSON(200, newList)
}
