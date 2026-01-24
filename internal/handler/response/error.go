package response

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteErrorResponse(w http.ResponseWriter, status int, code string, msg string) {
	var errResp ErrorResponse
	errResp.Error.Code = code
	errResp.Error.Message = msg
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
