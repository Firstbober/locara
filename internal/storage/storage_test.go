package storage

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Firstbober/locara/internal/models"
)

func TestGenerateNextID(t *testing.T) {
	tmpDir := t.TempDir()

	id, err := GenerateNextID(tmpDir)
	if err != nil {
		t.Fatalf("GenerateNextID() failed: %v", err)
	}

	if id != 1 {
		t.Errorf("GenerateNextID() = %d, want %d", id, 1)
	}

	archiveDir := archivePath(tmpDir, 1)
	if err := os.Mkdir(archiveDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	id, err = GenerateNextID(tmpDir)
	if err != nil {
		t.Fatalf("GenerateNextID() failed: %v", err)
	}

	if id != 2 {
		t.Errorf("GenerateNextID() = %d, want %d", id, 2)
	}
}

func TestSaveAndGetArchive(t *testing.T) {
	tmpDir := t.TempDir()

	meta := &models.Archive{
		Uploader:   "testuser",
		FileName:   "test.txt",
		SizeBytes:  100,
		MD5Sum:     "abc123",
		UploadedOn: time.Now(),
		Name:       "Test Archive",
		DatedOn:    "2024-01-01",
		Type:       "archive",
		Author:     "Test Author",
	}

	tmpFile := filepath.Join(tmpDir, "upload.txt")
	content := []byte("test content")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	header := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     int64(len(content)),
	}

	if err := SaveArchive(tmpDir, file, header, meta); err != nil {
		t.Fatalf("SaveArchive() failed: %v", err)
	}

	retrieved, err := GetArchive(tmpDir, 1)
	if err != nil {
		t.Fatalf("GetArchive() failed: %v", err)
	}

	if retrieved.ID != 1 {
		t.Errorf("GetArchive().ID = %d, want %d", retrieved.ID, 1)
	}

	if retrieved.Name != "Test Archive" {
		t.Errorf("GetArchive().Name = %s, want %s", retrieved.Name, "Test Archive")
	}

	filePath, err := GetArchiveFilePath(tmpDir, 1)
	if err != nil {
		t.Fatalf("GetArchiveFilePath() failed: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Archive file does not exist: %s", filePath)
	}
}

func TestListArchives(t *testing.T) {
	tmpDir := t.TempDir()

	for i := 1; i <= 3; i++ {
		archiveDir := archivePath(tmpDir, i)
		if err := os.Mkdir(archiveDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		meta := &models.Archive{
			ID:         i,
			Name:       fmt.Sprintf("Archive %d", i),
			Uploader:   "testuser",
			FileName:   fmt.Sprintf("file%d.txt", i),
			SizeBytes:  int64(i * 100),
			MD5Sum:     fmt.Sprintf("md5%d", i),
			UploadedOn: time.Now(),
			DatedOn:    "2024-01-01",
			Type:       "archive",
			Author:     "Test Author",
		}

		infoPath := infoFilePath(archiveDir)
		if err := writeInfoFile(infoPath, meta); err != nil {
			t.Fatalf("Failed to write info file: %v", err)
		}
	}

	archives, err := ListArchives(tmpDir)
	if err != nil {
		t.Fatalf("ListArchives() failed: %v", err)
	}

	if len(archives) != 3 {
		t.Errorf("ListArchives() returned %d archives, want %d", len(archives), 3)
	}
}
