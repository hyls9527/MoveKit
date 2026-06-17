package main

import "time"

type FolderInfo struct {
	Name       string `json:"name"`
	GUID       string `json:"guid"`
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	SizeBytes  int64  `json:"sizeBytes"`
	FileCount  int    `json:"fileCount"`
	Status     string `json:"status"` // "default" | "migrated" | "junction_broken"
}

type DiskInfo struct {
	Drive       string `json:"drive"`
	Label       string `json:"label"`
	TotalBytes  int64  `json:"totalBytes"`
	FreeBytes   int64  `json:"freeBytes"`
	IsSSD       bool   `json:"isSSD"`
	IsRemovable bool   `json:"isRemovable"`
	IsNetwork   bool   `json:"isNetwork"`
	FileSystem  string `json:"fileSystem"`
}

type DurationEstimate struct {
	IsCrossVolume bool   `json:"isCrossVolume"`
	TargetIsSSD   bool   `json:"targetIsSSD"`
	EstimatedSec  int    `json:"estimatedSec"`
	EstimatedText string `json:"estimatedText"`
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
	IsActive    bool     `json:"isActive"`
	Folders     []string `json:"folders"`
	HasGPOBlock bool     `json:"hasGPOBlock"`
}

type IntegrityIssue struct {
	Type    string `json:"type"`
	Folder  string `json:"folder"`
	Message string `json:"message"`
}

type ConflictInfo struct {
	RelativePath string `json:"relativePath"`
	SrcSize      int64  `json:"srcSize"`
	DstSize      int64  `json:"dstSize"`
	SrcModTime   string `json:"srcModTime"`
	DstModTime   string `json:"dstModTime"`
}

type CloudProvider struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type BackupRecord struct {
	Version       int       `json:"version"`
	Folder        string    `json:"folder"`
	GUID          string    `json:"guid"`
	OriginalPath  string    `json:"originalPath"`
	NewPath       string    `json:"newPath"`
	RegistryValue string    `json:"registryValue"`
	CreatedAt     time.Time `json:"createdAt"`
	Status        string    `json:"status"`
}

type MigrationState struct {
	Status     string    `json:"status"`
	Folder     string    `json:"folder"`
	Source     string    `json:"source"`
	Target     string    `json:"target"`
	StartTime  time.Time `json:"startTime"`
	FilesMoved int64     `json:"filesMoved"`
}

type MoveKitError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Action  string `json:"action"`
}

func (e *MoveKitError) Error() string { return e.Message }

type UpdateInfo struct {
	HasUpdate   bool   `json:"hasUpdate"`
	NewVersion  string `json:"newVersion"`
	DownloadURL string `json:"downloadURL"`
}

type MigrationStatus struct {
	State   string        `json:"state"`
	Folder  string        `json:"folder"`
	Percent int           `json:"percent"`
	Error   *MoveKitError `json:"error"`
}

type DiskDetail struct {
	Drive      string `json:"drive"`
	Label      string `json:"label"`
	IsSSD      bool   `json:"isSSD"`
	FreeBytes  int64  `json:"freeBytes"`
	TotalBytes int64  `json:"totalBytes"`
}
