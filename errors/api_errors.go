package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErr(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

func APIError(w http.ResponseWriter, err *ErrorResponse) {
	w.WriteHeader(err.Code)
	json.NewEncoder(w).Encode(err)
}

var (
	ErrBadRequestBody      = NewErr(400, "Bad Request body")
	ErrClientNotFound      = NewErr(404, "Client Not Found")
	ErrMethodNotAllowed    = NewErr(405, "Method not allowed")
	ErrRateLimitExceeded   = NewErr(429, "Rate limit exceeded")
	ErrClientNotUpdated    = NewErr(500, "Failed update client")
	ErrClientNotDeleted    = NewErr(500, "Failed delete client")
	ErrInternalServer      = NewErr(500, "An unexpected error occurred while processing the request")
	ErrServiceUnavailabled = NewErr(503, "Service unavailabled")
)
