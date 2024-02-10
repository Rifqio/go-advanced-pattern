package data

type Response struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"statusCode"`
	Result     any    `json:"result"`
	Message    string `json:"message"`
}

func NewResponse() Response {
	return Response{
		Status:     true,
		StatusCode: 200,
		Result:     nil,
		Message:    "",
	}
}
