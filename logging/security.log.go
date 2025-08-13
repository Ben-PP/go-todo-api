package logging

import "log/slog"

type SecurityScore int

const (
	SecurityScoreLow      SecurityScore = 1
	SecurityScoreMedium   SecurityScore = 5
	SecurityScoreHigh     SecurityScore = 10
	SecurityScoreCritical SecurityScore = 15
)

type SecurityEventName int

const (
	SecurityEventFailedLogin SecurityEventName = iota
	SecurityEventForbiddenAction
	SecurityEventJwtSignatureInvalid
	SecurityEventJwtReuse
	SecurityEventJwtUserUnknown
	SecurityEventJwtUnknown
	SecurityEventLoginToUnknownUsername
)

func (s SecurityEventName) String() string {
	switch s {
	case SecurityEventFailedLogin:
		return "failed-login"
	case SecurityEventJwtSignatureInvalid:
		return "jwt-signature-invalid-use"
	case SecurityEventJwtUserUnknown:
		return "jwt-user-not-found"
	case SecurityEventLoginToUnknownUsername:
		return "login-to-unknown-username"
	case SecurityEventJwtReuse:
		return "jwt-reuse"
	case SecurityEventJwtUnknown:
		return "jwt-unknown"
	}
	return "unknown"
}

func LogSecurityEvent(
	score SecurityScore,
	eventName SecurityEventName,
	targetPath string,
	target string,
	violator string,
) {
	log(
		slog.LevelInfo,
		"Security event has happened",
		"security",
		slog.Int("score", int(score)),
		slog.String("target_path", targetPath),
		slog.Group(
			"event",
			slog.String("name", eventName.String()),
			slog.String("target", target),
			slog.String("violator", violator),
		),
	)
}
