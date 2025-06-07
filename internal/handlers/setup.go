package handlers

import (
	"net/http"
	"strings"

	"bookify/internal/db"
	"bookify/internal/services"
	"bookify/internal/templates"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	DB    *db.Service
	Drive *services.DriveService
}

func render(c echo.Context, template templ.Component) error {
	return template.Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) IndexPage(c echo.Context) error {
	accounts, err := h.DB.ListAccounts()
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return c.Redirect(http.StatusFound, "/setup")
	}

	jobs, err := h.DB.ListRecentJobs(50)
	if err != nil {
		return err
	}

	return render(c, templates.MainPage(accounts, jobs))
}

func (h *Handlers) SetupPage(c echo.Context) error {
	// Get service account email for display
	serviceAccountEmail := h.Drive.GetServiceAccountEmail()
	return render(c, templates.SetupPage(serviceAccountEmail))
}

func (h *Handlers) CreateAccount(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	folderID := strings.TrimSpace(c.FormValue("folder_id"))
	serviceAccountEmail := h.Drive.GetServiceAccountEmail()

	if name == "" || folderID == "" {
		return render(c, templates.SetupPageWithError("All fields are required", serviceAccountEmail))
	}

	// Test if we can access the folder with the service account
	err := h.Drive.TestFolderAccess(folderID)
	if err != nil {
		return render(c, templates.SetupPageWithError("Cannot access folder. Make sure you've shared it with the service account email: "+serviceAccountEmail+". Error: "+err.Error(), serviceAccountEmail))
	}

	_, err = h.DB.CreateAccount(name, folderID)
	if err != nil {
		return render(c, templates.SetupPageWithError("Failed to create account: "+err.Error(), serviceAccountEmail))
	}

	// Use HTMX redirect header instead of HTTP redirect
	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}
