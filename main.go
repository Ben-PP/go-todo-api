package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
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

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
)

//go:embed db/migrations/*.sql
var migrationFiles embed.FS
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

	dbUrl := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d",
		config.Db.User,
		config.Db.Password,
		config.Db.Host,
		config.Db.Port,
	)
	if !config.Db.EnableSSL {
		dbUrl = fmt.Sprintf("%s?sslmode=disable", dbUrl)
	}

	migrationsDriver, err := iofs.New(migrationFiles, "db/migrations")
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logging.LogError(err, fmt.Sprintf("%v: %d", file, line), "Failed to initialize migration source.")
		return
	}

	migrate, err := migrate.NewWithSourceInstance(
		"iofs",
		migrationsDriver,
		dbUrl,
	)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logging.LogError(err, fmt.Sprintf("%v: %d", file, line), "Failed to initialize database migrations.")
		return
	}

	if err := migrate.Up(); err != nil {
		_, file, line, _ := runtime.Caller(1)
		logging.LogError(err, fmt.Sprintf("%v: %d", file, line), "Failed to run database migrations.")
		return
	}

	conn, err := pgx.Connect(context.Background(), dbUrl)
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

	if os.Getenv("GO_ENV") != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
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

	addr := fmt.Sprintf("%v:%v", config.Host, config.Port)
	if len(config.Proxies) > 0 {
		router.SetTrustedProxies(config.Proxies)
	}
	slog.Info("Using proxies: " + fmt.Sprint(config.Proxies))
	if config.SSL.Enabled {
		slog.Info("SSL enabled, starting HTTPS server.")
		router.RunTLS(addr, config.SSL.CertFile, config.SSL.KeyFile)
		return
	} else {
		router.Run(addr)
	}
}
