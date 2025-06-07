package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"bookify/internal/templates"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) UploadHandler(c echo.Context) error {
	accountIDStr := c.FormValue("account_id")
	if accountIDStr == "" {
		return render(c, templates.UploadError("Account ID is required"))
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		return render(c, templates.UploadError("Invalid account ID"))
	}

	account, err := h.DB.GetAccount(uint(accountID))
	if err != nil {
		return render(c, templates.UploadError("Account not found"))
	}

	form, err := c.MultipartForm()
	if err != nil {
		return render(c, templates.UploadError("Failed to parse form"))
	}

	files := form.File["files"]
	if len(files) == 0 {
		return render(c, templates.UploadError("No files provided"))
	}

	var jobIDs []string
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return render(c, templates.UploadError("Failed to create temp directory"))
	}

	for _, file := range files {
		if !isValidEPUB(file.Filename) {
			continue
		}

		src, err := file.Open()
		if err != nil {
			continue
		}
		defer func() {
			_ = src.Close() // Error ignored in cleanup
		}()

		tempPath := filepath.Join(tempDir, file.Filename)
		dst, err := os.Create(tempPath)
		if err != nil {
			continue
		}

		_, err = io.Copy(dst, src)
		_ = dst.Close() // Error ignored in cleanup
		if err != nil {
			_ = os.Remove(tempPath) // Error ignored
			continue
		}

		if !validateEPUBMagicBytes(tempPath) {
			_ = os.Remove(tempPath) // Error ignored
			continue
		}

		job, err := h.DB.CreateJob(account.ID, file.Filename)
		if err == nil {
			jobIDs = append(jobIDs, job.ID)
		}
	}

	if len(jobIDs) == 0 {
		return render(c, templates.UploadError("No valid EPUB files were uploaded"))
	}

	return render(c, templates.UploadSuccess(fmt.Sprintf("Successfully queued %d files for processing", len(jobIDs))))
}

func (h *Handlers) QueueStatusAPI(c echo.Context) error {
	jobs, err := h.DB.ListRecentJobs(50)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	html := ""
	for _, job := range jobs {
		component := templates.JobCard(job)
		var buf []byte
		if err := component.Render(c.Request().Context(), &writableBytes{&buf}); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to render template",
			})
		}
		html += string(buf)
	}

	return c.HTML(http.StatusOK, html)
}

func (h *Handlers) JobStatusAPI(c echo.Context) error {
	jobID := c.Param("id")
	job, err := h.DB.GetJob(jobID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Job not found",
		})
	}

	return c.JSON(http.StatusOK, job)
}

func isValidEPUB(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".epub"
}

func validateEPUBMagicBytes(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer func() {
		_ = file.Close() // Error ignored in validation
	}()

	header := make([]byte, 4)
	_, err = file.Read(header)
	if err != nil {
		return false
	}

	return header[0] == 0x50 && header[1] == 0x4B
}

type writableBytes struct {
	bytes *[]byte
}

func (w *writableBytes) Write(p []byte) (n int, err error) {
	*w.bytes = append(*w.bytes, p...)
	return len(p), nil
}
