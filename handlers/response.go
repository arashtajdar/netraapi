package handlers

import (
	"encoding/json"
	"net/http"
)

// APIError represents the structured error payload.
type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// RespondJSON serializes the data structure to a JSON response.
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondError serializes a standardized JSON error message.
// If the code is empty, it outputs a flat {"error": "message"} shape for backwards compatibility.
func RespondError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if code == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
