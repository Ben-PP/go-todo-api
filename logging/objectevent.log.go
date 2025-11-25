package logging

import (
	"log/slog"

	db "go-todo/db/sqlc"
)

type ObjectEvent int

const (
	ObjectEventCreate ObjectEvent = iota
	ObjectEventRead
	ObjectEventUpdate
	ObjectEventDelete
)

func (e ObjectEvent) String() string {
	switch e {
	case ObjectEventCreate:
		return "objectevent:create"
	case ObjectEventRead:
		return "objectevent:read"
	case ObjectEventUpdate:
		return "objectevent:update"
	case ObjectEventDelete:
		return "objectevent:delete"
	}
	return "objectevent:unknown"
}

// Logs object crud events. target of the event is identified by objectID and objectType.
func LogObjectEvent(
	targetPath string,
	srcIp string,
	eventType ObjectEvent,
	actor *db.User,
	objectID string,
	objectType string,
) {
	getActorData := func(id, username string, isAdmin bool) slog.Attr {
		return slog.Group(
			"actor",
			slog.String("id", id),
			slog.String("username", username),
			slog.Bool("is_admin", isAdmin),
		)
	}

	var actorData slog.Attr
	if actor != nil {
		actorData = getActorData(actor.ID, actor.Username, actor.IsAdmin)
	} else {
		actorData = getActorData("nil", "guest", false)
	}
	LogAuditEvent(
		true,
		targetPath,
		srcIp,
		eventType.String(),
		actorData,
		slog.Group(
			"object",
			slog.String("objecttype", objectType),
			slog.String("id", objectID),
		),
	)
}
