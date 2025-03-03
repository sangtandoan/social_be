package utils

type apiReponse struct {
	Data    any    `json:"data,omitempty"`
	Msg     string `json:"msg,omitempty"`
	Success bool   `json:"success,omitempty"`
}

func NewApiResponse(msg string, data any) *apiReponse {
	return &apiReponse{Msg: msg, Data: data}
}
