package schemas

type CreateShare struct {
	UserName string `json:"username" binding:"required"`
}

type Share struct {
	ID     string `json:"id"`
	ListID string `json:"list_id"`
	UserID string `json:"user_id"`
}

type ResponseShare struct {
	Share
	Username string `json:"username"`
}
