package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"bookify/internal/db"
	"bookify/internal/services"
	"bookify/internal/templates"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	DB      *db.Service
	Drive   *services.DriveService
	TempDir string
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
	return render(c, templates.SetupPage())
}

func (h *Handlers) CreateAccount(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	folderID := strings.TrimSpace(c.FormValue("folder_id"))

	if name == "" || folderID == "" {
		return render(c, templates.SetupPageWithError("All fields are required"))
	}

	// Check if account name already exists
	_, err := h.DB.GetAccountByName(name)
	if err == nil {
		return render(c, templates.SetupPageWithError("An account with this name already exists"))
	}

	// Redirect to OAuth flow with account setup parameters
	// URL encode the parameters
	params := make(url.Values)
	params.Set("account_name", name)
	params.Set("folder_id", folderID)
	redirectURL := "/oauth/start?" + params.Encode()

	c.Response().Header().Set("HX-Redirect", redirectURL)
	return c.NoContent(http.StatusOK)
}
