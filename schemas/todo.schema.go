package schemas

import "time"

type CreateTodo struct {
	Title          string     `json:"title" binding:"required"`
	Description    *string    `json:"description"`
	CompleteBefore *time.Time `json:"complete_before"`
	ParentID       *string    `json:"parent_id"`
}

type UpdateTodo struct {
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	CompleteBefore *time.Time `json:"complete_before"`
	Completed      *bool      `json:"completed"`
}
