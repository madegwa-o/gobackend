package main

import (
	"fmt"
	"gobackend/handlers"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/pay", handlers.HandlePayment)
	http.HandleFunc("/api/status", handlers.HandleStatus) // new GET endpoint

	fmt.Println("🚀 Server is  running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
