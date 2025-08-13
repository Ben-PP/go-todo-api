package user

import (
	"go-todo/middleware"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userController *UserController
}

func NewRoutes(userController *UserController) *UserRoutes {
	return &UserRoutes{userController}
}

func (routes *UserRoutes) Register(rg *gin.RouterGroup) {
	router := rg.Group("/user")
	router.GET("/:userID", middleware.JwtAuthMiddleware(), routes.userController.ReadUser)
	router.POST("/", routes.userController.CreateUser)
	router.PATCH("/:id", middleware.JwtAuthMiddleware(), routes.userController.UpdateUser)
	router.DELETE("/:id", middleware.JwtAuthMiddleware(), routes.userController.DeleteUser)
}
