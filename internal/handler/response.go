package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":  message,
		"status": http.StatusText(statusCode),
	}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
