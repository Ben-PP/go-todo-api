package todo

import (
	"errors"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (controller *TodoController) ReadShares(ctx *gin.Context) {
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

	list, err := controller.db.GetList(ctx, listID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
			return
		}
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get list", file, line, err, ctx)
		return
	}

	if list.UserID != reqUser.ID && !reqUser.IsAdmin {
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

	shares, err := controller.db.ReadShares(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to read shares", file, line, err, ctx)
		return
	}

	responseShares := make([]schemas.ResponseShare, 0, len(shares))
	for _, share := range shares {
		responseShares = append(responseShares, schemas.ResponseShare{
			ListID:   share.ListID,
			UserID:   share.UserID,
			Username: share.Username,
		})
	}

	// TODO Add logging. Add after object event logging has been reworked.
	ctx.JSON(200, responseShares)
}
