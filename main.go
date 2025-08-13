package main

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	db "go-todo/db/sqlc"
	"go-todo/features/auth"
	"go-todo/features/todo"
	"go-todo/features/user"
	"go-todo/logging"
	"go-todo/middleware"
	"go-todo/util/config"

	"github.com/jackc/pgx/v5"
)

var ctx context.Context

func main() {
	appLogger := logging.GetLogger()
	slog.SetDefault(appLogger)

	config, err := config.Get()
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logging.LogError(err, fmt.Sprintf("%v: %d", file, line), "Failed to load config.")
		return
	}

	conn, err := pgx.Connect(context.Background(), config.DbUrl)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logging.LogError(err, fmt.Sprintf("%v: %d", file, line), "Failed to connect to database.")
		return
	} else {
		fmt.Println("Connected to database")
	}

	defer conn.Close(ctx)

	mydb := db.New(conn)

	authController := auth.NewController(mydb, ctx)
	authRoutes := auth.NewRoutes(authController)
	userController := user.NewController(mydb, ctx)
	userRoutes := user.NewRoutes(userController)
	listController := todo.NewController(mydb, ctx)
	listRoutes := todo.NewRoutes(listController)

	router := gin.Default()

	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	{
		v1 := router.Group("/api/v1")
		v1.GET("/status", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"status": "ok"})
		})
		authRoutes.Register(v1)
		userRoutes.Register(v1)
		listRoutes.Register(v1)
	}

	slog.Info("Starting server.")
	router.Run(fmt.Sprintf("%v:8000", config.Host))
}
