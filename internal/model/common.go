package model

type CommonResponse struct {
	Prompts string `json:"prompts"`
	Status  int32  `json:"status"`
	Message string `json:"message"`
	Data    bool   `json:"data"`
}
