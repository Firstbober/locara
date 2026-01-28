package models

import "time"

// Archive represents archive metadata stored in info.json files.
type Archive struct {
	ID         int       `json:"id"`
	Uploader   string    `json:"uploader"`
	FileName   string    `json:"file_name"`
	SizeBytes  int64     `json:"size_bytes"`
	MD5Sum     string    `json:"md5_sum"`
	UploadedOn time.Time `json:"uploaded_on"`
	Name       string    `json:"name"`
	DatedOn    string    `json:"dated_on"`
	Type       string    `json:"type"`
	Author     string    `json:"author"`
}
