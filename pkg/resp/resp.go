package resp

type StdResponse struct {
	Prompts string      `json:"prompts"`
	Status  int32       `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewStdResponse(p IErrorCode, data interface{}) *StdResponse {
	return &StdResponse{
		Prompts: p.Prompts(),
		Message: p.Message(),
		Status:  p.Status(),
		Data:    data,
	}
}
