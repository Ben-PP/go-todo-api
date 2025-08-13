package schemas

import "time"

type CreateUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IsAdmin  bool   `json:"is_admin"`
}

type UpdateUser struct {
	Username string `json:"username" binding:"required"`
	IsAdmin  *bool  `json:"is_admin" binding:"required"`
}

type ResponseUser struct {
	Id        string    `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}
