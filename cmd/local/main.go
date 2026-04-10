package main

import (
	"context"
	"encoding/json"
	"lang-tracker/internal/db"
	"lang-tracker/internal/handler"
	"lang-tracker/internal/models"
	"lang-tracker/internal/service"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func tableName() string {
	if name := os.Getenv("TABLE_NAME"); name != "" {
		return name
	}
	return "lang-tracker"
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, relying on environment variables")
	}

	ctx := context.Background()
	client, err := db.New(ctx)
	if err != nil {
		log.Fatalf("failed to init DynamoDB client: %v", err)
	}

	table := tableName()
	svc := &handler.Services{
		Log:   &service.LogService{DB: client, TableName: table},
		Stats: &service.StatsService{DB: client, TableName: table},
	}

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit request body to 1 MB to prevent abuse
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req models.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		apiReq := handler.ToAPIGatewayRequest(req)
		resp, err := svc.Handler(r.Context(), apiReq)
		if err != nil {
			slog.ErrorContext(r.Context(), "handler returned unexpected error", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write([]byte(resp.Body))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("local server starting", "addr", "http://localhost:"+port+"/api")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
