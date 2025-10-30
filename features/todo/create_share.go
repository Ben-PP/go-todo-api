package todo

import (
	"errors"
	"fmt"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Creates a share for a list
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

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventCreate,
		&reqUser,
		&listShare,
		nil,
		logging.ObjectEventSubListShare,
	)
	ctx.JSON(201, gin.H{"status": "shared", "list": listShare.ListID})
}
