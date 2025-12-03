package user

import (
	"errors"
	"net/http"
	"reflect"
	"runtime"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"
	"go-todo/logging"
	"go-todo/schemas"
	"go-todo/util/mycontext"
	"go-todo/util/passwd"
	"go-todo/util/validate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser handles user creation requests.
//
//	@Summary		Create user
//	@Description	Creates a new user account. The first user created is granted admin privileges.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			user	body		schemas.CreateUser	true	"User creation payload"
//	@Success		201		{object}	db.User				"Returns the created user object."
//	@Failure		400		{object}	schemas.ErrorResponse	"Bad request due to invalid input."
//	@Failure		500		{object}	schemas.ErrorResponse	"Internal server error."
//	@Router			/user [post]
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
		user.ID,
		reflect.TypeOf(db.User{}).String(),
	)
	ctx.JSON(http.StatusCreated, user)
}
