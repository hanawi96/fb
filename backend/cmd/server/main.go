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
	apiRouter.HandleFunc("/pages/unassigned", handler.GetUnassignedPages).Methods("GET")
	apiRouter.HandleFunc("/pages/{id}", handler.DeletePage).Methods("DELETE")
	apiRouter.HandleFunc("/pages/{id}/toggle", handler.TogglePage).Methods("PATCH")
	apiRouter.HandleFunc("/pages/{id}/assignments", handler.GetPageAssignments).Methods("GET")
	apiRouter.HandleFunc("/pages/{id}/assign", handler.AssignPageToAccount).Methods("POST")
	apiRouter.HandleFunc("/pages/{id}/assign/{accountId}", handler.UnassignPageFromAccount).Methods("DELETE")
	apiRouter.HandleFunc("/pages/{id}/primary", handler.SetPrimaryAccount).Methods("PUT")
	apiRouter.HandleFunc("/pages/{id}/timeslots", handler.GetPageTimeSlots).Methods("GET")
	apiRouter.HandleFunc("/pages/{id}/timeslots", handler.CreateTimeSlot).Methods("POST")
	
	// Posts routes
	apiRouter.HandleFunc("/posts", handler.CreatePost).Methods("POST")
	apiRouter.HandleFunc("/posts", handler.GetPosts).Methods("GET")
	apiRouter.HandleFunc("/posts/publish", handler.PublishPost).Methods("POST")
	apiRouter.HandleFunc("/posts/{id}", handler.GetPost).Methods("GET")
	apiRouter.HandleFunc("/posts/{id}", handler.UpdatePost).Methods("PUT")
	apiRouter.HandleFunc("/posts/{id}", handler.DeletePost).Methods("DELETE")
	
	// Schedule routes
	apiRouter.HandleFunc("/schedule", handler.SchedulePost).Methods("POST")
	apiRouter.HandleFunc("/schedule", handler.GetScheduledPosts).Methods("GET")
	apiRouter.HandleFunc("/schedule/preview", handler.PreviewSchedule).Methods("POST")
	apiRouter.HandleFunc("/schedule/smart", handler.ScheduleWithPreview).Methods("POST")
	apiRouter.HandleFunc("/schedule/stats", handler.GetScheduleStats).Methods("GET")
	apiRouter.HandleFunc("/schedule/{id}", handler.DeleteScheduledPost).Methods("DELETE")
	apiRouter.HandleFunc("/schedule/{id}/retry", handler.RetryScheduledPost).Methods("POST")
	
	// Logs routes
	apiRouter.HandleFunc("/logs", handler.GetPostLogs).Methods("GET")
	
	// Hashtag routes
	apiRouter.HandleFunc("/hashtags/search", handler.SearchHashtags).Methods("GET")
	apiRouter.HandleFunc("/hashtags/saved", handler.GetSavedHashtags).Methods("GET")
	apiRouter.HandleFunc("/hashtags/saved", handler.SaveHashtags).Methods("POST")
	apiRouter.HandleFunc("/hashtags/saved", handler.DeleteSavedHashtag).Methods("DELETE")

	// Facebook Accounts routes (Multi-Account System)
	apiRouter.HandleFunc("/accounts", handler.GetAccounts).Methods("GET")
	apiRouter.HandleFunc("/accounts", handler.CreateAccount).Methods("POST")
	apiRouter.HandleFunc("/accounts/{id}", handler.GetAccount).Methods("GET")
	apiRouter.HandleFunc("/accounts/{id}", handler.UpdateAccount).Methods("PUT")
	apiRouter.HandleFunc("/accounts/{id}", handler.DeleteAccount).Methods("DELETE")
	apiRouter.HandleFunc("/accounts/{id}/pages", handler.GetAccountPages).Methods("GET")
	apiRouter.HandleFunc("/accounts/{id}/refresh", handler.RefreshAccountToken).Methods("POST")

	// Notifications routes
	apiRouter.HandleFunc("/notifications", handler.GetNotifications).Methods("GET")
	apiRouter.HandleFunc("/notifications/count", handler.GetUnreadCount).Methods("GET")
	apiRouter.HandleFunc("/notifications/read-all", handler.MarkAllNotificationsRead).Methods("PUT")
	apiRouter.HandleFunc("/notifications/{id}/read", handler.MarkNotificationRead).Methods("PUT")
	apiRouter.HandleFunc("/notifications/{id}", handler.DeleteNotification).Methods("DELETE")

	// Time Slots routes
	apiRouter.HandleFunc("/timeslots/{id}", handler.UpdateTimeSlot).Methods("PUT")
	apiRouter.HandleFunc("/timeslots/{id}", handler.DeleteTimeSlot).Methods("DELETE")
	
	// Upload route
	apiRouter.HandleFunc("/upload", handler.UploadImage).Methods("POST")
	
	// Serve uploaded files
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

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
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Tăng lên 2 phút cho upload nhiều ảnh
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 Server running on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
