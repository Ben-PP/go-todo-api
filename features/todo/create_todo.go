package todo

import (
	"errors"
	"runtime"
	"slices"
	"time"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"
	"go-todo/util/validate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (controller *TodoController) CreateTodo(ctx *gin.Context) {
	payload := &schemas.CreateTodo{}
	description := ""

	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}
	requesterId, requesterUsername, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}
	listID := ctx.Param("listID")

	if ok := validate.LengthTitle(payload.Title); !ok {
		ctx.Error(gterrors.NewGtValueError(payload.Title, "title too long"))
		return
	}
	if payload.Description != nil {
		if ok := validate.LengthDescription(*payload.Description); !ok {
			ctx.Error(gterrors.NewGtValueError(*payload.Description, "description too long"))
			return
		}
		description = *payload.Description
	}

	// Check users right to access the list
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
			listID,
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	parentID := ""
	if payload.ParentID != nil {
		parentID = *payload.ParentID
	}
	var completeBefore time.Time
	if payload.CompleteBefore != nil {
		completeBefore = *payload.CompleteBefore
	} else {
		completeBefore = time.Date(1970, 0o1, 0o1, 0o0, 0o0, 0o0, 0o0, time.UTC)
	}

	args := &db.CreateTodoParams{
		ID:             uuid.New().String(),
		ListID:         listID,
		UserID:         reqUser.ID,
		Title:          payload.Title,
		Description:    pgtype.Text{String: description, Valid: payload.Description != nil},
		ParentID:       pgtype.Text{String: parentID, Valid: payload.ParentID != nil},
		CompleteBefore: pgtype.Timestamp{Time: completeBefore, Valid: payload.CompleteBefore != nil},
	}

	todo, err := controller.db.CreateTodo(ctx, *args)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to create todo",
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
		logging.ObjectEventCreate,
		reqUser,
		&todo,
		nil,
		logging.ObjectEventSubTodo,
	)
	ctx.JSON(201, gin.H{"status": "created", "todo": todo})
}
