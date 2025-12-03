package user

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/database"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// ReadUser handles requests to read user information.
//
//	@Summary		Read user information
//	@Description	Retrieves user information by user ID. Accessible by the user themselves or an admin.
//	@Security		Bearer
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string					true	"User ID"
//	@Success		200		{object}	schemas.ResponseUser	"Returns user information."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		401		{object}	schemas.ErrorResponse	"Unauthorized due to invalid or expired token."
//	@Failure		403		{object}	schemas.ErrorResponse	"Forbidden action."
//	@Failure		404		{object}	schemas.ErrorResponse	"User not found."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/user/{userID} [get]
func (controller *UserController) ReadUser(ctx *gin.Context) {
	requesterId, requesterUsername, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	userIDToGet := ctx.Param("userID")

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

	if reqUser.ID != userIDToGet && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreMedium,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			fmt.Sprintf("userID: %v", userIDToGet),
			reqUser.ID,
		)
		return
	}

	user, err := controller.db.GetUserById(ctx, userIDToGet)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
		}
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError(
			"failed to get user from db",
			file,
			line,
			err,
			ctx,
		)
		return
	}

	responseUser := &schemas.ResponseUser{
		Id:        user.ID,
		Username:  user.Username,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.Time,
	}

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventRead,
		reqUser,
		user.ID,
		reflect.TypeOf(user).String(),
	)
	ctx.JSON(200, *responseUser)
}
