package logging

import (
	"log/slog"
)

type SessionEventType int

const (
	SessionEventTypeLogin SessionEventType = iota
	SessionEventTypeLogout
	SessionEventTypeRefresh
)

func (s SessionEventType) String() string {
	switch s {
	case SessionEventTypeLogin:
		return "session:login"
	case SessionEventTypeLogout:
		return "session:logout"
	case SessionEventTypeRefresh:
		return "session:refresh"
	}
	return "unknown"
}

// Logs user session events.
func LogSessionEvent(
	success bool,
	targetPath string,
	username string,
	eventType SessionEventType,
	srcIp string,
) {
	LogAuditEvent(
		success,
		targetPath,
		srcIp,
		eventType.String(),
		slog.Group(
			"target",
			slog.String("username", username),
		),
	)
}
