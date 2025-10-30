package schemas

type CreateShare struct {
	UserName string `json:"username" binding:"required"`
}
