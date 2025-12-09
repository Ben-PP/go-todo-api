package todo

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Removes a share for a list
//
//	@Summary		Remove list share
//	@Description	Removes a share for a todo list by its ID. Only the owner, the shared user, or an admin can remove the share.
//	@Tags			Shares
//	@Security		Bearer
//	@Produce		json
//	@Param			listID	path		string	true	"ID of the todo list to unshare"
//	@Param			shareID	path		string	true	"ID of the share to remove"
//	@Success		204		"No content, indicating successful unsharing."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to missing or invalid JWT."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action for non-owners/non-admins/non-shared-users."
//	@Failure		404		{object}	schemas.ErrorResponse	"Todo list or share not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/list/{listID}/share/{shareID} [delete]
func (controller *TodoController) DeleteShare(ctx *gin.Context) {
	listID := ctx.Param("listID")
	shareID := ctx.Param("shareID")

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
	listShare, err := controller.db.GetShareById(ctx, shareID)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
			return
		}
		mycontext.CtxAddGtInternalError("failed to get list share", file, line, err, ctx)
		return
	}

	if list.UserID != reqUser.ID && (!reqUser.IsAdmin && reqUser.ID != listShare.UserID) {
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

	err = controller.db.DeleteShare(ctx, shareID)
	if err != nil {
		var pgErr *pgconn.PgError
		errMessage := "failed to unshare list"
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
		logging.ObjectEventDelete,
		&reqUser,
		shareID,
		reflect.TypeOf(listShare).String(),
	)
	ctx.JSON(204, gin.H{})
}
