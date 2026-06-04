package handlers

import (
	"encoding/json"
	"net/http"
)

// writeError writes a consistent JSON error response: {"error":"message"}
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
