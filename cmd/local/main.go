package main

import (
	"encoding/json"
	"fmt"
	"lang-tracker/internal/db"
	"lang-tracker/internal/handler"
	"lang-tracker/internal/models"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Simple HTTP adapter to mimic API Gateway locally
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found")
	}

	db.Client = db.NewFakeDynamo()

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Wrap into an API Gateway-like request
		apiReq := handler.ToAPIGatewayRequest(req)
		resp, _ := handler.Handler(r.Context(), apiReq)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write([]byte(resp.Body))
	})

	port := getPort()
	fmt.Printf("🚀 Local server running at http://localhost:%s/api\n", port)
	_ = http.ListenAndServe(":"+port, nil)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
