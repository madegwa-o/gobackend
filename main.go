package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// PaymentRequest represents the incoming JSON body
type PaymentRequest struct {
	Amount   float64 `json:"amount"`
	Phone    string  `json:"phone"`
	Ref      string  `json:"ref"`
	Callback string  `json:"callback"`
}

// PaymentResponse is what we return to the client
type PaymentResponse struct {
	Status  string  `json:"status"`
	Message string  `json:"message"`
	Amount  float64 `json:"amount"`
}

func main() {
	http.HandleFunc("/api/pay", handlePayment)
	http.HandleFunc("/api/status", handleStatus) // new GET endpoint

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /api/pay — receive a payment request
func handlePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("✅ Received payment request: %+v\n", req)

	res := PaymentResponse{
		Status:  "success",
		Message: "Payment request received successfully",
		Amount:  req.Amount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// GET /api/status — simple status check
func handleStatus(w http.ResponseWriter, r *http.Request) {
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
