package main

import (
	"encoding/json"
	"net/http"
)

// APIError struktur untuk error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RespondError mengirim error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	error := APIError{
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(error)
}

// RespondSuccess mengirim success response
func respondSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(data)
}