package main

import (
	"chirpy-project/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		platform = "dev"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	jwt_secret := os.Getenv("JWT_SECRET")

	polka_key := os.Getenv("POLKA_KEY")

	const filepathRoot = "."
	const port = "8080"

	mux := http.NewServeMux()

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
		jwtSecret:      jwt_secret,
		polkaKey:       polka_key,
	}
	// Initialize apiConfig

	mux.HandleFunc("GET /api/healthz", healthzHandler) // Register healthzHandler for /healthz path
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler) // Added cfg. to validateChirpHandler
	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	mux.HandleFunc("GET /api/chirps", cfg.listChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpid}", cfg.getChirpHandler)
	mux.HandleFunc("POST /api/login", cfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", cfg.revokeRefreshTokenHandler)
	mux.HandleFunc("PUT /api/users", cfg.updateUserHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpid}", cfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.upgradeUserHandler)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
