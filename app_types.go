package main

import "time"

type FolderInfo struct {
	Name       string `json:"name"`
	GUID       string `json:"guid"`
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	Status     string `json:"status"`    // "default" | "migrated" | "junction_broken"
	SizeBytes  int64  `json:"sizeBytes"` // -1 means "could not calculate"
	FileCount  int    `json:"fileCount"`
}

type DiskInfo struct {
	Drive        string `json:"drive"`
	Label        string `json:"label"`
	FileSystem   string `json:"fileSystem"`
	TotalBytes   int64  `json:"totalBytes"`
	FreeBytes    int64  `json:"freeBytes"`
	IsFixedDrive bool   `json:"isFixedDrive"`
	IsRemovable  bool   `json:"isRemovable"`
	IsNetwork    bool   `json:"isNetwork"`
}

type DurationEstimate struct {
	EstimatedText      string `json:"estimatedText"`
	EstimatedSec       int    `json:"estimatedSec"`
	IsCrossVolume      bool   `json:"isCrossVolume"`
	TargetIsFixedDrive bool   `json:"targetIsFixedDrive"`
}

type ProgressInfo struct {
	Folder     string `json:"folder"`
	Phase      string `json:"phase"`
	FilesMoved int    `json:"filesMoved"`
	TotalFiles int    `json:"totalFiles"`
	BytesMoved int64  `json:"bytesMoved"`
	TotalBytes int64  `json:"totalBytes"`
	Percent    int    `json:"percent"`
}

type OneDriveStatus struct {
	Folders     []string `json:"folders"`
	IsActive    bool     `json:"isActive"`
	HasGPOBlock bool     `json:"hasGPOBlock"`
}

type IntegrityIssue struct {
	Type    string `json:"type"`
	Folder  string `json:"folder"`
	Message string `json:"message"`
}

type ConflictInfo struct {
	RelativePath string `json:"relativePath"`
	SrcModTime   string `json:"srcModTime"`
	DstModTime   string `json:"dstModTime"`
	SrcSize      int64  `json:"srcSize"`
	DstSize      int64  `json:"dstSize"`
}

type CloudProvider struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type BackupRecord struct {
	CreatedAt     time.Time `json:"createdAt"`
	Folder        string    `json:"folder"`
	GUID          string    `json:"guid"`
	OriginalPath  string    `json:"originalPath"`
	NewPath       string    `json:"newPath"`
	RegistryValue string    `json:"registryValue"`
	Status        string    `json:"status"`
	Version       int       `json:"version"`
}

type MigrationState struct {
	StartTime  time.Time `json:"startTime"`
	Status     string    `json:"status"`
	Folder     string    `json:"folder"`
	Source     string    `json:"source"`
	Target     string    `json:"target"`
	FilesMoved int64     `json:"filesMoved"`
}

type MoveKitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Action  string `json:"action"`
	Warning string `json:"warning,omitempty"` // Non-blocking informational warning
}

func (e *MoveKitError) Error() string { return e.Message }

type UpdateInfo struct {
	NewVersion  string `json:"newVersion"`
	DownloadURL string `json:"downloadURL"`
	HasUpdate   bool   `json:"hasUpdate"`
}

type MigrationStatus struct {
	Error   *MoveKitError `json:"error"`
	State   string        `json:"state"`
	Folder  string        `json:"folder"`
	Percent int           `json:"percent"`
}

type DiskDetail struct {
	Drive        string `json:"drive"`
	Label        string `json:"label"`
	IsFixedDrive bool   `json:"isFixedDrive"`
	FreeBytes    int64  `json:"freeBytes"`
	TotalBytes   int64  `json:"totalBytes"`
}
