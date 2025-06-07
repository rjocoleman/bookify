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
	driveService := services.NewDriveService()

	// Test service account configuration
	log.Println("Testing service account configuration...")
	err = driveService.TestConnection()
	if err != nil {
		log.Printf("WARNING: Service account test failed: %v", err)
		log.Println("Make sure GOOGLE_SERVICE_ACCOUNT_KEY_PATH or GOOGLE_SERVICE_ACCOUNT_KEY is set")
	} else {
		log.Println("✓ Service account authentication successful")
	}

	queueService := services.NewQueueService(dbService, driveService)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	handlers := &handlers.Handlers{
		DB:    dbService,
		Drive: driveService,
	}

	e.GET("/", handlers.IndexPage)
	e.GET("/setup", handlers.SetupPage)
	e.POST("/setup", handlers.CreateAccount)
	e.POST("/upload", handlers.UploadHandler)
	e.GET("/api/queue", handlers.QueueStatusAPI)
	e.GET("/api/job/:id", handlers.JobStatusAPI)

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
