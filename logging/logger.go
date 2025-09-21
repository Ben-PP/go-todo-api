package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/DeRuina/timberjack"
)

func GetLogger() *slog.Logger {
	logDir := "/var/log/go-todo"
	if os.Getenv("GO_ENV") == "dev" {
		logDir = "."
	}
	_, err := os.Stat(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Log directory does not exist.")
		} else {
			panic("Unknown error while testing log directory.")
		}
	}
	testFile := fmt.Sprintf("%s/test.log", logDir)
	f, err := os.Create(testFile)
	if err != nil {
		panic(fmt.Sprintf("Could not write the log directory '%s'", logDir))
	}
	f.Close()
	os.Remove(testFile)

	logRotator := &timberjack.Logger{
		Filename:   fmt.Sprintf("%s/app.log", logDir),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     100,
		Compress:   true,
	}

	appMultiWriter := slog.NewJSONHandler(
		io.MultiWriter(logRotator),
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		},
	)

	appLogger := slog.New(appMultiWriter).With(slog.String("program_name", "GO-TODO"))

	return appLogger
}
