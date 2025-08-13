package user

import (
	"context"
	db "go-todo/db/sqlc"
)

type UserController struct {
	db  *db.Queries
	ctx context.Context
}

func NewController(db *db.Queries, ctx context.Context) *UserController {
	return &UserController{db: db, ctx: ctx}
}
