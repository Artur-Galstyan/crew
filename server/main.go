package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	apiKey string
}

func authMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			clientApiKey := strings.TrimPrefix(authHeader, "Bearer ")
			if clientApiKey == "" || clientApiKey != apiKey {
				slog.Warn("Unauthorized access attempt",
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path,
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Successfully accessed protected register route"))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}

	crewApiKey := os.Getenv("CREW_API_KEY")
	cfg := config{
		apiKey: crewApiKey,
	}

	slog.Info("config is", "config", cfg)

	router := http.NewServeMux()
	router.HandleFunc("/health", health)
	router.Handle("POST /register", authMiddleware(cfg.apiKey)(http.HandlerFunc(register)))

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	slog.Info("Server running on port 8080")
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
