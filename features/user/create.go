package user

import (
	"errors"
	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/mycontext"
	"go-todo/util/passwd"
	"go-todo/util/validate"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func (controller *UserController) CreateUser(ctx *gin.Context) {
	var payload *schemas.CreateUser
	if ok := mycontext.ShouldBindBodyWithJSON(&payload, ctx); !ok {
		return
	}

	makeAdmin := false
	users, err := controller.db.GetAllUsers(ctx)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to get users from db", file, line, err, ctx)
		return
	}
	if len(users) == 0 {
		makeAdmin = true
	}

	isPasswdValid, err := validate.Password(payload.Password)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to validate new password", file, line, err, ctx)
		return
	} else if !isPasswdValid {
		ctx.Error(gterrors.ErrPasswordUnsatisfied).SetType(gin.ErrorTypePublic)
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

	userUUID := uuid.New()
	password := payload.Password
	passwdHash, err := passwd.Hash(password)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		mycontext.CtxAddGtInternalError("failed to hash new password", file, line, err, ctx)
		return
	}

	args := &db.CreateUserParams{
		ID:           userUUID.String(),
		Username:     payload.Username,
		PasswordHash: passwdHash,
		IsAdmin:      makeAdmin,
	}

	user, err := controller.db.CreateUser(ctx, *args)
	if err != nil {
		var pgErr *pgconn.PgError
		errMessage := "failed to create user"
		if errors.As(err, &pgErr) {
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

	logging.LogObjectEvent(
		ctx.FullPath(),
		ctx.ClientIP(),
		logging.ObjectEventCreate,
		nil,
		&user,
		nil,
		logging.ObjectEventSubUser,
	)
	ctx.JSON(http.StatusCreated, gin.H{"status": "created", "user": user})
}
