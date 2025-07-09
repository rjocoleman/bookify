package main

import (
	"log"
	"os"

	"bookify/internal/db"
	"bookify/internal/handlers"
	"bookify/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./kepub.db"
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	dbService := db.NewService(database)
	driveService := services.NewDriveService(dbService)

	// Check OAuth configuration
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Println("WARNING: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set for OAuth authentication")
	} else {
		log.Println("✓ OAuth configuration found")
	}

	queueService := services.NewQueueService(dbService, driveService)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	h := &handlers.Handlers{
		DB:    dbService,
		Drive: driveService,
	}

	oauthHandlers := handlers.NewOAuthHandlers(dbService)

	e.GET("/", h.IndexPage)
	e.GET("/setup", h.SetupPage)
	e.POST("/setup", h.CreateAccount)
	e.POST("/upload", h.UploadHandler)
	e.GET("/api/queue", h.QueueStatusAPI)
	e.GET("/api/job/:id", h.JobStatusAPI)

	// OAuth routes
	e.GET("/oauth/start", oauthHandlers.StartOAuth)
	e.GET("/oauth/callback", oauthHandlers.OAuthCallback)

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "./temp"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Warning: Failed to create temp directory: %v", err)
	}

	go queueService.StartWorker()
	go queueService.StartCleanupWorker()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(e.Start(":" + port))
}
