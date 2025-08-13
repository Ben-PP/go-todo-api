package auth

import (
	"go-todo/middleware"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authController *AuthController
}

func NewRoutes(authController *AuthController) *AuthRoutes {
	return &AuthRoutes{authController}
}

func (routes *AuthRoutes) Register(rg *gin.RouterGroup) {
	router := rg.Group("/auth")
	router.POST("/login", routes.authController.Login)
	router.POST("/logout", middleware.JwtAuthMiddleware(), routes.authController.Logout)
	router.POST("/refresh", routes.authController.Refresh)
	router.POST("/update-password", middleware.JwtAuthMiddleware(), routes.authController.UpdatePassword)
}
