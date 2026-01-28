package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Firstbober/locara/internal/models"
)

//go:embed templates
var templateFS embed.FS

// ParseTemplates parses HTML templates from embedded filesystem.
func ParseTemplates() (*template.Template, error) {
	return ParseTemplatesWithFS(templateFS)
}

// ParseTemplatesFromFS tries to parse from filesystem first, then falls back to embed.
func ParseTemplatesFromFS() (*template.Template, error) {
	// Try filesystem first (for development)
	tmpl := template.New("").Funcs(getTemplateFuncMap())

	// Parse main templates
	mainGlob := "internal/templates/templates/*.html"
	mainFiles, err := filepath.Glob(mainGlob)
	if err == nil {
		for _, file := range mainFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			name := filepath.Base(file)
			if _, err := tmpl.New(name).Parse(string(content)); err != nil {
				fmt.Printf("Warning: failed to parse template %s: %v\n", file, err)
			}
		}
	}

	// Parse partials
	partialsGlob := "internal/templates/templates/partials/*.html"
	partialsFiles, err := filepath.Glob(partialsGlob)
	if err == nil {
		for _, file := range partialsFiles {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			name := "partials/" + filepath.Base(file)
			if _, err := tmpl.New(name).Parse(string(content)); err != nil {
				fmt.Printf("Warning: failed to parse template %s: %v\n", file, err)
			}
		}
	}

	if len(mainFiles) > 0 || len(partialsFiles) > 0 {
		return tmpl, nil
	}

	// Fallback to embedded templates
	return ParseTemplates()
}

// ParseTemplatesWithFS parses templates from a given filesystem.
func ParseTemplatesWithFS(fsys fs.FS) (*template.Template, error) {
	tmpl := template.New("").Funcs(getTemplateFuncMap())

	err := fs.WalkDir(fsys, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !isTemplateFile(path) {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		name := fsPathToTemplateName(path)
		if _, err := tmpl.New(name).Parse(string(content)); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// getTemplateFuncMap returns custom template functions.
func getTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"prettyBytes": prettyBytes,
		"formatDate":  formatDate,
		"groupByYear": groupByYear,
	}
}

// prettyBytes formats bytes into human-readable string.
func prettyBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatDate formats a date string for display.
func formatDate(s string) string {
	return s
}

func isTemplateFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".html" || ext == ".htm"
}

func fsPathToTemplateName(path string) string {
	path = strings.TrimPrefix(path, "templates/")
	path = strings.ReplaceAll(path, "/", ":")
	return path
}

// groupByYear groups archives by year based on DatedOn field.
func groupByYear(archives []models.Archive) map[int][]models.Archive {
	yearMap := make(map[int][]models.Archive)

	for _, archive := range archives {
		parsedTime, err := time.Parse("2006-01-02", archive.DatedOn)
		if err != nil {
			continue
		}
		year := parsedTime.Year()
		yearMap[year] = append(yearMap[year], archive)
	}

	// Sort years in descending order for display
	years := make([]int, 0, len(yearMap))
	for year := range yearMap {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	// Return ordered map
	result := make(map[int][]models.Archive)
	for _, year := range years {
		result[year] = yearMap[year]
	}

	return result
}
