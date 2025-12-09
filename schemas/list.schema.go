package schemas

import db "go-todo/db/sqlc"

type CreateList struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
}

type UpdateList struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type ListWithTodos struct {
	db.List
	Todos []db.Todo `json:"todos"`
}
