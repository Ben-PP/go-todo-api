package schemas

type ErrorResponse struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}
