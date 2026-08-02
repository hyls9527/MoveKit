// Hand-written mirror of the Wails-generated bindings (frontend/wailsjs/go/models.ts)
// used for the no-Wails dev/mock fallback. Keep field names in sync with the Go
// types in app_types.go when adding or renaming fields.

export interface FolderInfo {
  name: string;
  guid: string;
  sourcePath: string;
  targetPath: string;
  sizeBytes: number;
  fileCount: number;
  status: string; // "default" | "migrated" | "not_on_c" | "junction_broken" | "unknown"
}

export interface DiskInfo {
  drive: string;
  label: string;
  totalBytes: number;
  freeBytes: number;
  isFixedDrive: boolean;
  isRemovable: boolean;
  isNetwork: boolean;
  fileSystem: string;
}

export interface ProgressInfo {
  folder: string;
  phase: string;
  filesMoved: number;
  totalFiles: number;
  bytesMoved: number;
  totalBytes: number;
  percent: number;
}

export type AppState = 'idle' | 'checking' | 'confirming' | 'migrating' | 'completed' | 'failed';
