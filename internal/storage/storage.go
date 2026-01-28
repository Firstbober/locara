package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Firstbober/locara/internal/models"
)

const (
	infoFileName = "info.json"
)

// SaveArchive saves an uploaded file and its metadata to a new archive directory.
func SaveArchive(baseDir string, file io.Reader, header *multipart.FileHeader, meta *models.Archive) error {
	newID, err := GenerateNextID(baseDir)
	if err != nil {
		return fmt.Errorf("failed to generate next ID: %w", err)
	}

	archiveDir := archivePath(baseDir, newID)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	meta.ID = newID

	infoPath := infoFilePath(archiveDir)
	if err := writeInfoFile(infoPath, meta); err != nil {
		return fmt.Errorf("failed to write info file: %w", err)
	}

	filePath := filepath.Join(archiveDir, header.Filename)
	if err := saveFile(filePath, file); err != nil {
		return fmt.Errorf("failed to save archive file: %w", err)
	}

	return nil
}

// GetArchive reads and returns the archive metadata for the given ID.
func GetArchive(baseDir string, id int) (*models.Archive, error) {
	archiveDir := archivePath(baseDir, id)
	infoPath := infoFilePath(archiveDir)

	data, err := os.ReadFile(infoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read info file: %w", err)
	}

	var archive models.Archive
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, fmt.Errorf("failed to parse info file: %w", err)
	}

	return &archive, nil
}

// ListArchives returns all archives in the uploads directory.
func ListArchives(baseDir string) ([]models.Archive, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploads directory: %w", err)
	}

	var archives []models.Archive

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id, err := parseArchiveID(entry.Name())
		if err != nil {
			continue
		}

		archive, err := GetArchive(baseDir, id)
		if err != nil {
			continue
		}

		archives = append(archives, *archive)
	}

	return archives, nil
}

// GenerateNextID finds the highest existing archive ID and returns the next one.
func GenerateNextID(baseDir string) (int, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return 1, nil
	}

	highestID := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id, err := parseArchiveID(entry.Name())
		if err != nil {
			continue
		}

		if id > highestID {
			highestID = id
		}
	}

	return highestID + 1, nil
}

// GetArchiveFilePath returns the full path to the archive file for the given ID.
func GetArchiveFilePath(baseDir string, id int) (string, error) {
	archive, err := GetArchive(baseDir, id)
	if err != nil {
		return "", fmt.Errorf("failed to get archive metadata: %w", err)
	}

	filePath := filepath.Join(archivePath(baseDir, id), archive.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("archive file does not exist: %s", filePath)
	}

	return filePath, nil
}

func archivePath(baseDir string, id int) string {
	return filepath.Join(baseDir, strconv.Itoa(id))
}

func infoFilePath(archiveDir string) string {
	return filepath.Join(archiveDir, infoFileName)
}

func parseArchiveID(dirName string) (int, error) {
	return strconv.Atoi(dirName)
}

func writeInfoFile(path string, meta *models.Archive) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func saveFile(path string, src io.Reader) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
