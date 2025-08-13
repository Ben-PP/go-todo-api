package logging

import (
	"io"
	"log/slog"

	"github.com/DeRuina/timberjack"
)

func GetLogger() *slog.Logger {
	logRotator := &timberjack.Logger{
		Filename:   "./app.log",
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
