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

// ReadShares handles reading shares for a specific todo list.
//
//	@Summary		Get shares for a todo list
//	@Description	Retrieves all shares associated with a specific todo list. Requires ownership of the list or admin privileges.
//	@Tags			Shares
//	@Security		Bearer
//	@Produce		json
//	@Param			listID	path		string	true	"ID of the todo list"
//	@Success		200		{array}		schemas.ResponseShare	"Returns a list of shares for the specified todo list."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for non-owners or non-admins."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID}/share [get]
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

	shareRows, err := controller.db.ReadShares(ctx, listID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to read shares", file, line, err, ctx)
		return
	}

	shares := make([]schemas.ResponseShare, 0, len(shareRows))
	for _, share := range shareRows {
		shares = append(shares, schemas.ResponseShare{
			Share: schemas.Share{
				ID:     share.ID,
				ListID: share.ListID,
				UserID: share.UserID,
			},
			Username: share.Username,
		})
	}

	// TODO Add logging. Add after object event logging has been reworked.
	ctx.JSON(200, shares)
}
