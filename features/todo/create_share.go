package todo

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Creates a share for a list
//
//	@Summary		Share a todo list with another user
//	@Description	Shares a todo list with another user by their username.
//	@Tags			Shares
//	@Security		Bearer
//	@Accept			json
//	@Produce		json
//	@Param			listID	path		string					true	"ID of the todo list to share"
//	@Param			share	body		schemas.CreateShare		true	"Share details"
//	@Success		201		{object}	schemas.ResponseShare	"Returns the share ID upon successful sharing."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list or user not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID}/share [post]
func (controller *TodoController) CreateShare(ctx *gin.Context) {
	var payload *schemas.CreateShare
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}
	listID := ctx.Param("listID")

	tokenUserId, tokenUserName, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	reqUser, err := controller.db.GetUserById(ctx, tokenUserId)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			tokenUserName,
			ctx.ClientIP(),
		)
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonJwtUserNotFound,
				fmt.Errorf("could not get user from db: %w", err),
			),
		).SetType(gterrors.GetGinErrorType())
		return
	}

	list, err := controller.db.GetList(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
			return
		}
		mycontext.CtxAddGtInternalError("failed to get list", file, line, err, ctx)
		return
	}

	if list.UserID != reqUser.ID && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreMedium,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			list.ID,
			reqUser.ID,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	userToAdd, err := controller.db.GetUserByUsername(ctx, payload.UserName)
	if err != nil {
		logging.LogSecurityEvent(
			logging.SecurityScoreLow,
			logging.SecurityEventJwtUserUnknown,
			ctx.FullPath(),
			payload.UserName,
			ctx.ClientIP(),
		)

		ctx.Error(
			gterrors.NewGtValueError(
				payload.UserName,
				"Username not found",
			),
		).SetType(gin.ErrorTypePublic)
		return
	}

	args := &db.CreateShareParams{
		ID:     uuid.New().String(),
		ListID: listID,
		UserID: userToAdd.ID,
	}

	listShare, err := controller.db.CreateShare(ctx, *args)
	if err != nil {
		var pgErr *pgconn.PgError
		errMessage := "failed to share a list"
		if errors.As(err, &pgErr) {
			_, file, line, _ := runtime.Caller(0)
			switch pgErr.Code {
			case "23505":
				mycontext.CtxAddGtInternalError(
					"failed to create unique id for list",
					file,
					line,
					err,
					ctx,
				)
			default:
				_, file, line, _ := runtime.Caller(0)
				mycontext.CtxAddGtInternalError(errMessage, file, line, err, ctx)
			}
			return
		}
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(errMessage, file, line, err, ctx)
		return
	}

	share := &schemas.ResponseShare{
		Share: schemas.Share{
			ID:     listShare.ID,
			ListID: listShare.ListID,
			UserID: listShare.UserID,
		},
		Username: payload.UserName,
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventCreate,
		&reqUser,
		listShare.ID,
		reflect.TypeOf(db.ListShare{}).String(),
	)
	ctx.JSON(201, *share)
}
