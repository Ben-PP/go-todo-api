package database

import (
	"fmt"

	db "go-todo/db/sqlc"
	"go-todo/gterrors"

	"github.com/gin-gonic/gin"
)

// Gets the user for id from db and handles any errors with gin. Return user or
// error which is already pushed to gin.Context
func GetUserById(db *db.Queries, id string, ctx *gin.Context) (*db.User, error) {
	user, err := db.GetUserById(ctx, id)
	if err != nil {
		ctx.Error(
			gterrors.NewGtAuthError(
				gterrors.GtAuthErrorReasonJwtUserNotFound,
				fmt.Errorf("could not get user from db: %w", err),
			),
		).SetType(gterrors.GetGinErrorType())
		return nil, err
	}
	return &user, nil
}
