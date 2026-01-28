package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Firstbober/locara/internal/config"
	"github.com/Firstbober/locara/internal/models"
	"github.com/Firstbober/locara/internal/storage"
)

// CreateArchiveHandler handles file uploads with metadata.
func CreateArchiveHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("[ERROR] Failed to parse multipart form: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	authCode := r.FormValue("ar_auth_code")
	if authCode == "" {
		log.Printf("[ERROR] Missing auth code")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if !validateAuthCode(cfg, authCode) {
		log.Printf("[ERROR] Invalid auth code")
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("ar_file")
	if err != nil {
		log.Printf("[ERROR] Failed to get uploaded file: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer file.Close()

	var uploader string
	for _, user := range cfg.Users {
		if user.Auth == authCode {
			uploader = user.Name
			break
		}
	}

	meta := &models.Archive{
		Uploader:   uploader,
		FileName:   header.Filename,
		SizeBytes:  header.Size,
		MD5Sum:     header.Header.Get("Content-MD5"),
		UploadedOn: time.Now(),
		Name:       r.FormValue("ar_name"),
		DatedOn:    r.FormValue("ar_dated"),
		Type:       r.FormValue("ar_type"),
		Author:     r.FormValue("ar_author"),
	}

	if meta.Name == "" || meta.DatedOn == "" || meta.Type == "" || meta.Author == "" {
		log.Printf("[ERROR] Missing required fields")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := storage.SaveArchive(cfg.UseDirectory, file, header, meta); err != nil {
		log.Printf("[ERROR] Failed to save archive: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("[INFO] Archive created: ID=%d, Name=%s", meta.ID, meta.Name)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ListArchivesHandler returns a JSON list of all archives.
func ListArchivesHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	archives, err := storage.ListArchives(cfg.UseDirectory)
	if err != nil {
		log.Printf("[ERROR] Failed to list archives: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to list archives")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(archives); err != nil {
		log.Printf("[ERROR] Failed to encode archives: %v", err)
	}
}

// DownloadArchiveHandler handles file downloads.
func DownloadArchiveHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	idStr := r.PathValue("id")
	if idStr == "" {
		log.Printf("[ERROR] Missing archive ID")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		log.Printf("[ERROR] Invalid archive ID: %s", idStr)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	filePath, err := storage.GetArchiveFilePath(cfg.UseDirectory, id)
	if err != nil {
		log.Printf("[ERROR] Failed to get archive file: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	archive, err := storage.GetArchive(cfg.UseDirectory, id)
	if err != nil {
		log.Printf("[ERROR] Failed to get archive metadata: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to open file: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", archive.FileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("[ERROR] Failed to send file: %v", err)
		return
	}

	log.Printf("[INFO] Archive downloaded: ID=%d", id)
}
