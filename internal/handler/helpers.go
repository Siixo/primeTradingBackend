package handler

import (
	"encoding/json"
	"net/http"
)

// jsonError sends a consistent JSON-formatted error response.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
