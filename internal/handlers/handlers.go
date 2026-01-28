package handlers

import (
	"html/template"
	"net/http"

	"github.com/Firstbober/locara/internal/config"
)

// validateAuthCode checks if the provided auth code matches any configured user.
func validateAuthCode(cfg *config.Config, code string) bool {
	for _, user := range cfg.Users {
		if user.Auth == code {
			return true
		}
	}
	return false
}

// writeJSONError writes an HTTP error response as JSON.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// renderTemplate renders an HTML template with the given data.
func renderTemplate(w http.ResponseWriter, tmpl *template.Template, name string, data any) error {
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		return err
	}
	return nil
}
