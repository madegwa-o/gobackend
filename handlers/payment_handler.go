package handlers

import (
	"encoding/json"
	"fmt"
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

func HandlePayment(W http.ResponseWriter, R *http.Request) {
	if R.Method != http.MethodPost {
		http.Error(W, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var Req PaymentRequest
	err := json.NewDecoder(R.Body).Decode(&Req)
	if err != nil {
		http.Error(W, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("✅ Received payment request: %+v\n", Req)

	res := PaymentResponse{
		Status:  "success",
		Message: "Payment request received successfully",
		Amount:  Req.Amount,
	}

	W.Header().Set("Content-Type", "application/json")
	json.NewEncoder(W).Encode(res)
}
