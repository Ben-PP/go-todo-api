package todo

import (
	"context"
	db "go-todo/db/sqlc"
)

type TodoController struct {
	db  *db.Queries
	ctx context.Context
}

func NewController(db *db.Queries, ctx context.Context) *TodoController {
	return &TodoController{db: db, ctx: ctx}
}
