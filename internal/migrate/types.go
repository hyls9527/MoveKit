package migrate

import "time"

// ProgressInfo holds progress information during a file move operation.
// This mirrors the ProgressInfo from app_types.go but lives in the migrate
// package to avoid importing package main.
type ProgressInfo struct {
	Folder     string `json:"folder"`
	Phase      string `json:"phase"`
	FilesMoved int    `json:"filesMoved"`
	TotalFiles int    `json:"totalFiles"`
	BytesMoved int64  `json:"bytesMoved"`
	TotalBytes int64  `json:"totalBytes"`
	Percent    int    `json:"percent"`
}

// ConflictInfo describes a file that exists at both source and target.
type ConflictInfo struct {
	RelativePath string `json:"relativePath"`
	SrcModTime   string `json:"srcModTime"`
	DstModTime   string `json:"dstModTime"`
	SrcSize      int64  `json:"srcSize"`
	DstSize      int64  `json:"dstSize"`
}

// CloudProvider represents a detected cloud sync provider.
type CloudProvider struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

// MigrationState tracks the current state of an in-progress migration.
type MigrationState struct {
	StartTime  time.Time `json:"startTime"`
	Status     string    `json:"status"`
	Folder     string    `json:"folder"`
	Source     string    `json:"source"`
	Target     string    `json:"target"`
	FilesMoved int64     `json:"filesMoved"`
}

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	NewVersion  string `json:"newVersion"`
	DownloadURL string `json:"downloadURL"`
	HasUpdate   bool   `json:"hasUpdate"`
}
