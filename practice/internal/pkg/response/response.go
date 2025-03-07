package response

type ApiResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    any    `json:"data"`
}

func NewApiResponse(msg string, data any) *ApiResponse {
	return &ApiResponse{
		Success: true,
		Msg:     msg,
		Data:    data,
	}
}
