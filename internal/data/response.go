package data

type Response struct {
	Status     bool        `json:"status"`
	StatusCode int         `json:"statusCode"`
	Result     interface{} `json:"result"`
	Message    string      `json:"message"`
}

func NewResponse() Response {
	return Response{
		Status:     true,
		StatusCode: 200,
		Result:     nil,
		Message:    "",
	}
}
