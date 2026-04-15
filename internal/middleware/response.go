package middleware

import (
	"encoding/json"
	"net/http"
)

// writeJSONError sends a consistent JSON-formatted error response from middleware.
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
