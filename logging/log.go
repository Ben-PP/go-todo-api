package logging

import (
	"context"
	"log/slog"
)

func log(level slog.Leveler, message string, sourcetype string, args ...slog.Attr) {
	slog.LogAttrs(
		context.Background(),
		level.Level(),
		message,
		append(
			[]slog.Attr{slog.String("sourcetype", "go-todo:"+sourcetype)},
			args...,
		)...,
	)
}
