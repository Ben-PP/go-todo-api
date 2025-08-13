package logging

import "log/slog"

func LogError(err error, file string, detail string) {
	log(
		slog.LevelError,
		"Error happened during request handling.",
		"error",
		slog.String("detail", detail),
		slog.String("error", err.Error()),
		slog.String("file", file),
	)
}
