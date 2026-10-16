package response

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status"`
}
