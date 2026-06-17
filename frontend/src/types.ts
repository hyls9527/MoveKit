export interface FolderInfo {
  name: string;
  guid: string;
  sourcePath: string;
  targetPath: string;
  sizeBytes: number;
  fileCount: number;
  status: string; // "default" | "migrated" | "junction_broken"
}

export interface DiskInfo {
  drive: string;
  label: string;
  totalBytes: number;
  freeBytes: number;
  isSSD: boolean;
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
