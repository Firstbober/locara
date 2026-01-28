package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/Firstbober/locara/internal/config"
	"github.com/Firstbober/locara/internal/models"
	"github.com/Firstbober/locara/internal/storage"
)

// IndexHandler renders the main page with the archive list.
func IndexHandler(tmpl *template.Template, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		archives, err := storage.ListArchives(cfg.UseDirectory)
		if err != nil {
			log.Printf("[ERROR] Failed to list archives: %v", err)
			archives = []models.Archive{}
		}

		data := struct {
			Archives []models.Archive
			Cfg      *config.Config
		}{
			Archives: archives,
			Cfg:      cfg,
		}

		if err := renderTemplate(w, tmpl, "index.html", data); err != nil {
			log.Printf("[ERROR] Failed to render template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// UploadHandler renders the upload form page.
func UploadHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, tmpl, "upload.html", nil); err != nil {
			log.Printf("[ERROR] Failed to render template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// ErrorHandler renders an error page.
func ErrorHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := renderTemplate(w, tmpl, "error.html", nil); err != nil {
			log.Printf("[ERROR] Failed to render template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
