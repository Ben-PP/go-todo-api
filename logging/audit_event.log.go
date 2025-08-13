package logging

import "log/slog"

func LogAuditEvent(
	success bool,
	targetPath string,
	srcIp string,
	eventType string,
	args ...any,
) {
	log(
		slog.LevelInfo,
		"Event audited",
		"audit",
		slog.Group(
			"event",
			append(
				[]any{
					slog.String("eventtype", eventType),
					slog.Bool("success", success),
					slog.String("src_ip", srcIp),
					slog.String("target_path", targetPath),
				},
				args...,
			)...,
		),
	)
}
