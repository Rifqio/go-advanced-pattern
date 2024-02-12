package data

type Pagination struct {
	PageSize  int    `json:"page_size"`
	Page      int    `json:"page"`
	TotalData int    `json:"total_data"`
	Sort      string `json:"sort"`
}
type Response struct {
	Status     bool        `json:"status"`
	StatusCode int         `json:"statusCode"`
	Result     interface{} `json:"result"`
	Message    string      `json:"message"`
	Pagination Pagination  `json:"pagination"`
}

func NewResponse() Response {
	return Response{
		Status:     true,
		StatusCode: 200,
		Result:     nil,
		Message:    "",
	}
}
