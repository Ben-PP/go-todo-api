package mycontext

import (
	"errors"
	"fmt"

	"go-todo/gterrors"
	"go-todo/util/txtutil"

	"github.com/gin-gonic/gin"
)

func getBooleanKey(key string, c *gin.Context) (bool, bool) {
	var isAdmin bool
	isAdminRaw, exists := c.Get(key)
	if exists {
		if val, ok := isAdminRaw.(bool); ok {
			isAdmin = val
		} else {
			exists = false
		}
	}
	return isAdmin, exists
}

func GetTokenVariables(ctx *gin.Context) (userID string, username string, isAdmin bool, err error) {
	userID = ctx.GetString("x-token-user-id")
	username = ctx.GetString("x-token-username")
	isAdmin, isAdminExists := getBooleanKey("x-token-is-admin", ctx)

	if userID == "" || username == "" || !isAdminExists {
		return "", "", false, errors.New("failed to get token variables")
	}

	return userID, username, isAdmin, nil
}

func CtxAddGtInternalError(message, file string, line int, err error, c *gin.Context) {
	errToAdd := err
	if message != "" {
		errToAdd = fmt.Errorf("%v: %w", message, err)
	}
	c.Error(
		gterrors.NewGtInternalError(
			errToAdd,
			txtutil.AddLineNumberToFileName(file, line),
			500,
		),
	).SetType(gterrors.GetGinErrorType())
}

// Check the body format against the schema. 'payload' should be like
// &schemas.<MyBodySchema>
func ShouldBindBodyWithJSON(payload any, c *gin.Context) bool {
	if err := c.ShouldBindBodyWithJSON(payload); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return false
	}
	return true
}
