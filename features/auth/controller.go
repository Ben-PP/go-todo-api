package auth

import (
	"context"
	db "go-todo/db/sqlc"
	"go-todo/logging"
	"go-todo/util/jwt"
	"go-todo/util/mycontext"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	db  *db.Queries
	ctx context.Context
}

func NewController(db *db.Queries, ctx context.Context) *AuthController {
	return &AuthController{db: db, ctx: ctx}
}

func logTokenEventUse(success bool, token *jwt.GtClaims, c *gin.Context) {
	logging.LogTokenEvent(success, c.FullPath(), logging.TokenEventTypeUse, c.RemoteIP(), token)
}

func logTokenCreations(claims []*jwt.GtClaims, c *gin.Context) {
	for _, claims := range claims {
		logging.LogTokenEvent(
			true,
			c.FullPath(),
			logging.TokenEventtypeCreate,
			c.RemoteIP(),
			claims,
		)
	}
}

func generateTokens(family string, user db.User) (
	refreshToken string,
	refreshClaims *jwt.GtClaims,
	accessToken string,
	accessClaims *jwt.GtClaims,
	err error,
) {
	refreshToken, refreshClaims, err = jwt.GenerateRefreshJwt(
		user.Username,
		user.ID,
		user.IsAdmin,
		family,
	)
	if err != nil {
		return "", nil, "", nil, err
	}
	accessToken, accessClaims, err = jwt.GenerateAccessJwt(
		user.Username,
		user.ID,
		user.IsAdmin,
	)
	if err != nil {

		return "", nil, "", nil, err
	}
	return
}

func failedToGenerateJwtError(err error, file string, line int, c *gin.Context) {
	mycontext.CtxAddGtInternalError("failed to generate jwt", file, line, err, c)
}

func failedToSaveJwtToDbError(err error, file string, line int, c *gin.Context) {
	mycontext.CtxAddGtInternalError("failed to save jwt to db", file, line, err, c)
}

// TODO Add ResetPassword (Requires email to be implemented)
