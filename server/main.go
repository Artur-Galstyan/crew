package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver for GORM
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type config struct {
	apiKey string
}

type server struct {
	db  *gorm.DB
	cfg config
}

type User struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"size:255;not null;unique" json:"name"`
}

type RegisterRequest struct {
	Name string `json:"name"`
}

type LoginRequest struct {
	Name string `json:"name"`
}

func (s *server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if strings.TrimPrefix(authHeader, "Bearer ") != s.cfg.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("Failed to decode JSON", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newUser := User{
		Name: req.Name,
	}

	result := s.db.Create(&newUser)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created: " + req.Name))
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

	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Could not find home directory:", "err", err)
	}
	crewPath := filepath.Join(home, ".crew")
	err = os.MkdirAll(crewPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	sqliteDbPath := filepath.Join(crewPath, "crew.db")
	db, err := gorm.Open(sqlite.Open(sqliteDbPath), &gorm.Config{})

	err = db.AutoMigrate(&User{})
	if err != nil {
		slog.Error("Auto-migration failed!", "error", err)
		os.Exit(1)
	}

	srv := server{
		db:  db,
		cfg: cfg,
	}

	router := http.NewServeMux()
	router.HandleFunc("/health", health)
	router.Handle("POST /register", srv.auth(http.HandlerFunc(srv.register)))
	router.Handle("POST /login", srv.auth(http.HandlerFunc(srv.login)))

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	slog.Info("Server running on port 8080")
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
