package todo

import (
	"go-todo/middleware"

	"github.com/gin-gonic/gin"
)

type TodoRoutes struct {
	todoController *TodoController
}

func NewRoutes(todoController *TodoController) *TodoRoutes {
	return &TodoRoutes{todoController}
}

func (routes *TodoRoutes) Register(rg *gin.RouterGroup) {
	router := rg.Group("/list")

	router.Use(middleware.JwtAuthMiddleware())

	router.GET("/", routes.todoController.ReadLists)
	router.POST("/", routes.todoController.CreateList)
	router.GET("/:listID", routes.todoController.ReadListWithTodos)
	router.PATCH("/:listID", routes.todoController.UpdateList)
	router.DELETE("/:listID", routes.todoController.DeleteList)
	router.GET("/:listID/share", routes.todoController.ReadShares)
	router.POST("/:listID/share", routes.todoController.CreateShare)
	router.DELETE("/:listID/share/:shareID", routes.todoController.DeleteShare)

	todoRouter := router.Group("/:listID/todo")
	todoRouter.POST("/", routes.todoController.CreateTodo)
	todoRouter.PATCH("/:todoID", routes.todoController.UpdateTodo)
	todoRouter.DELETE("/:todoID", routes.todoController.DeleteTodo)

	// TODO Implement get shares
}
