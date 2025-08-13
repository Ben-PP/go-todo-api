package logging

import "log/slog"

func LogReq(src_ip string, method string, path string, userAgent string, status int) {
	log(
		slog.LevelInfo,
		"Request handled",
		"request",
		slog.Group(
			"request",
			slog.String("src_ip", src_ip),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("user_agent", userAgent),
		),
		slog.Group("response", slog.Int("status", status)),
	)
}
