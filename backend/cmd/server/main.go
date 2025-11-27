package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"fbscheduler/internal/api"
	"fbscheduler/internal/db"
	"fbscheduler/internal/scheduler"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	database, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Test connection
	if err := database.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// Initialize database store
	store := db.NewStore(database)

	// Initialize API handlers
	handler := api.NewHandler(store, database)

	// Setup router
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	// Login routes (no auth required)
	apiRouter.HandleFunc("/login", handler.Login).Methods("POST")
	apiRouter.HandleFunc("/verify", handler.VerifyToken).Methods("GET")
	
	// Auth routes
	apiRouter.HandleFunc("/auth/facebook/url", handler.GetFacebookAuthURL).Methods("GET")
	apiRouter.HandleFunc("/auth/facebook/callback", handler.FacebookCallback).Methods("POST")
	apiRouter.HandleFunc("/auth/pages/save", handler.SaveSelectedPages).Methods("POST")
	apiRouter.HandleFunc("/auth/debug/pages", handler.DebugPages).Methods("POST")
	
	// Pages routes
	apiRouter.HandleFunc("/pages", handler.GetPages).Methods("GET")
	apiRouter.HandleFunc("/pages/{id}", handler.DeletePage).Methods("DELETE")
	apiRouter.HandleFunc("/pages/{id}/toggle", handler.TogglePage).Methods("PATCH")
	
	// Posts routes
	apiRouter.HandleFunc("/posts", handler.CreatePost).Methods("POST")
	apiRouter.HandleFunc("/posts", handler.GetPosts).Methods("GET")
	apiRouter.HandleFunc("/posts/{id}", handler.GetPost).Methods("GET")
	apiRouter.HandleFunc("/posts/{id}", handler.UpdatePost).Methods("PUT")
	apiRouter.HandleFunc("/posts/{id}", handler.DeletePost).Methods("DELETE")
	
	// Schedule routes
	apiRouter.HandleFunc("/schedule", handler.SchedulePost).Methods("POST")
	apiRouter.HandleFunc("/schedule", handler.GetScheduledPosts).Methods("GET")
	apiRouter.HandleFunc("/schedule/{id}", handler.DeleteScheduledPost).Methods("DELETE")
	apiRouter.HandleFunc("/schedule/{id}/retry", handler.RetryScheduledPost).Methods("POST")
	
	// Logs routes
	apiRouter.HandleFunc("/logs", handler.GetPostLogs).Methods("GET")
	
	// Upload route
	apiRouter.HandleFunc("/upload", handler.UploadImage).Methods("POST")

	// CORS configuration
	allowedOrigins := []string{os.Getenv("FRONTEND_URL")}
	// Allow file:// protocol for testing
	if os.Getenv("ALLOW_FILE_PROTOCOL") == "true" {
		allowedOrigins = append(allowedOrigins, "null")
	}
	
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Start scheduler in background
	sched := scheduler.NewScheduler(store)
	go sched.Start()
	log.Println("✅ Scheduler started")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      c.Handler(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Server running on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
