package gterrors

import (
	"errors"
	"go-todo/util/config"

	"github.com/gin-gonic/gin"
)

var ErrForbidden = errors.New("forbidden")
var ErrJwtRefreshReuse = errors.New("refresh jwt reuse")
var ErrNotFound = errors.New("resource not found")
var ErrPasswordUnsatisfied = errors.New("password criteria not met")
var ErrPasswordSame = errors.New("password cannot be the old one")
var ErrShouldNotHappen = errors.New("this should not happen")
var ErrUniqueViolation = errors.New("already exists")
var ErrUsernameUnsatisfied = errors.New("username criteria not met")

// Returns either gin.ErrorTypePrivate or gin.ErrorTypePublic if GO_ENV is "dev".
func GetGinErrorType() gin.ErrorType {
	ginType := gin.ErrorTypePrivate
	if config.GetGoEnv() == "dev" {
		ginType = gin.ErrorTypePublic
	}
	return ginType
}
