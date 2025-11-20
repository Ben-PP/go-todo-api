package schemas

type CreateShare struct {
	UserName string `json:"username" binding:"required"`
}

type ResponseShare struct {
	ListID   string `json:"list_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
