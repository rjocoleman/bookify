package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bookify/internal/db"
	"bookify/internal/services"
	"bookify/internal/testutil"

	"github.com/labstack/echo/v4"
)

func TestHandlers_SetupPage_Simple(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := services.NewDriveService(dbService)

	handlers := &Handlers{
		DB:    dbService,
		Drive: driveService,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handlers.SetupPage(c)
	if err != nil {
		t.Errorf("SetupPage() error = %v", err)
	}

	testutil.AssertResponseStatus(t, rec, http.StatusOK)
	testutil.AssertResponseContains(t, rec, "Setup Bookify")
}

func TestHandlers_IndexPage_NoAccounts(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := services.NewDriveService(dbService)

	handlers := &Handlers{
		DB:    dbService,
		Drive: driveService,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handlers.IndexPage(c)
	if err != nil {
		t.Errorf("IndexPage() error = %v", err)
	}

	// Should redirect to setup when no accounts exist
	testutil.AssertResponseStatus(t, rec, http.StatusFound)

	location := rec.Header().Get("Location")
	if location != "/setup" {
		t.Errorf("Expected redirect to /setup, got: %q", location)
	}
}

func TestHandlers_IndexPage_WithAccounts(t *testing.T) {
	// Create test database
	testDB := testutil.SetupTestDB(t)
	err := testDB.AutoMigrate(&db.Account{}, &db.Job{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	dbService := db.NewService(testDB)
	driveService := services.NewDriveService(dbService)

	// Create a test account
	_, err = dbService.CreateAccount("Test Account", "folder123")
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	handlers := &Handlers{
		DB:    dbService,
		Drive: driveService,
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handlers.IndexPage(c)
	if err != nil {
		t.Errorf("IndexPage() error = %v", err)
	}

	// Should show the main page when accounts exist
	testutil.AssertResponseStatus(t, rec, http.StatusOK)
	testutil.AssertResponseContains(t, rec, "Bookify")
}
