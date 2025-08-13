package todo

import (
	"errors"
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
		&newList,
		&oldList,
		logging.ObjectEventSubList,
	)
	ctx.JSON(200, gin.H{"status": "ok", "list": newList})
}
