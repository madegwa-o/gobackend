package handlers

import (
	"encoding/json"
	"net/http"
)

func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]string{
		"status":  "running",
		"version": "1.0.0",
		"message": "Go payment gateway API is healthy 🚀",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
