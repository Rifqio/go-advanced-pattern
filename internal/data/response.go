package data

type Response struct {
	Status  bool   `json:"status"`
	Result  any    `json:"result"`
	Message string `json:"message"`
}
