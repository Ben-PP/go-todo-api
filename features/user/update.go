package user

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/mycontext"
	"go-todo/util/validate"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (controller *UserController) UpdateUser(ctx *gin.Context) {
	tokenUserId, _, _, err := mycontext.GetTokenVariables(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get claims from jwt", file, line, err, ctx)
		return
	}

	reqUser, err := controller.db.GetUserById(ctx, tokenUserId)
	if err != nil {
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonTokenInvalid,
				err,
			),
		).SetType(gterrors.GetGinErrorType())
		return
	}

	userIDToUpdate := ctx.Param("id")

	if userIDToUpdate != reqUser.ID && !reqUser.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreHigh,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			userIDToUpdate,
			reqUser.Username,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}
	var payload *schemas.UpdateUser
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}
	isUsernameValid, err := validate.Username(payload.Username)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to validate new username", file, line, err, ctx)
		return
	} else if !isUsernameValid {
		ctx.Error(gterrors.ErrUsernameUnsatisfied).SetType(gin.ErrorTypePublic)
		return
	}

	if !reqUser.IsAdmin && *payload.IsAdmin {
		logging.LogSecurityEvent(
			logging.SecurityScoreHigh,
			logging.SecurityEventForbiddenAction,
			ctx.FullPath(),
			userIDToUpdate,
			reqUser.Username,
		)
		ctx.Error(gterrors.ErrForbidden).SetType(gin.ErrorTypePublic)
		return
	}

	var oldUser *db.User
	if userIDToUpdate != reqUser.ID {
		userFromDB, err := controller.db.GetUserById(ctx, userIDToUpdate)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ctx.Error(gterrors.ErrNotFound).SetType(gin.ErrorTypePublic)
				return
			}
			_, file, line, _ := runtime.Caller(0)
			mycontext.CtxAddGtInternalError("could not get user from db", file, line, err, ctx)
			return
		}
		oldUser = &userFromDB
	} else {
		oldUser = &reqUser
	}
	if oldUser.Username == payload.Username && oldUser.IsAdmin == *payload.IsAdmin {
		logging.LogObjectEvent(
			ctx.FullPath(),
			ctx.ClientIP(),
			logging.ObjectEventUpdate,
			&reqUser,
			&oldUser,
			&oldUser,
			logging.ObjectEventSubUser,
		)
		ctx.JSON(http.StatusNoContent, gin.H{})
		return
	}

	args := &db.UpdateUserParams{
		ID:       userIDToUpdate,
		Username: payload.Username,
		IsAdmin:  *payload.IsAdmin,
	}

	updatedUser, err := controller.db.UpdateUser(ctx, *args)
	if err != nil {
		var pgErr *pgconn.PgError
		errMessage := "failed to update user"
		if errors.As(err, &pgErr) {
			fmt.Println("pgErr: ", pgErr)
			switch pgErr.Code {
			case "23505":
				ctx.Error(gterrors.ErrUniqueViolation).SetType(gin.ErrorTypePublic)
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

	if oldUser.IsAdmin != updatedUser.IsAdmin {
		if err := controller.db.DeleteJwtTokensByUserId(ctx, updatedUser.ID); err != nil {
			var pgErr *pgconn.PgError
			errMessage := "failed to delete old jwts"
			if errors.As(err, &pgErr) {
				fmt.Println("pgErr: ", pgErr)
				switch pgErr.Code {
				default:
					_, file, line, _ := runtime.Caller(0)
					logging.LogError(
						fmt.Errorf("%s: %w", errMessage, err),
						fmt.Sprintf("%v: %d", file, line),
						err.Error(),
					)
				}
				return
			}
			_, file, line, _ := runtime.Caller(0)
			logging.LogError(
				fmt.Errorf("%s: %w", errMessage, err),
				fmt.Sprintf("%v: %d", file, line),
				err.Error(),
			)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"user":   updatedUser,
	})
}
