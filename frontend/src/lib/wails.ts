// Wails bindings wrapper.
// In production (wails build), imports from the auto-generated bindings.
// In dev mode without Wails, falls back to mock data.

import type { FolderInfo, DiskInfo, ProgressInfo } from '../types';

// Try to import real Wails bindings; fall back to mocks if not available
let realBindings: any = null;
try {
  // @ts-ignore - wailsjs is generated at build time
  realBindings = await import('../../wailsjs/go/main/App');
} catch {
  // No real bindings available — use mocks
}

// ---- Mock fallbacks ----

const mockFolders: FolderInfo[] = [
  { name: '桌面', guid: '{B4BFCC3A}', sourcePath: 'C:\\Users\\User\\Desktop', targetPath: '', sizeBytes: 5.2e9, fileCount: 127, status: 'default' },
  { name: '下载', guid: '{374DE290}', sourcePath: 'C:\\Users\\User\\Downloads', targetPath: '', sizeBytes: 3.8e9, fileCount: 340, status: 'default' },
  { name: '音乐', guid: '{4BD8D571}', sourcePath: 'C:\\Users\\User\\Music', targetPath: '', sizeBytes: 0.1e9, fileCount: 12, status: 'default' },
  { name: '文档', guid: '{FDD39AD0}', sourcePath: 'C:\\Users\\User\\Documents', targetPath: '', sizeBytes: 1.2e9, fileCount: 89, status: 'default' },
  { name: '图片', guid: '{33E28130}', sourcePath: 'C:\\Users\\User\\Pictures', targetPath: '', sizeBytes: 2.4e9, fileCount: 156, status: 'default' },
];

const mockDisks: DiskInfo[] = [
  { drive: 'C:', label: '系统盘', totalBytes: 256e9, freeBytes: 45e9, isSSD: true, isRemovable: false, isNetwork: false, fileSystem: 'NTFS' },
  { drive: 'D:', label: '数据盘', totalBytes: 512e9, freeBytes: 128e9, isSSD: true, isRemovable: false, isNetwork: false, fileSystem: 'NTFS' },
];

// ---- Exported functions (delegate to real bindings or mock) ----

export async function GetFolders(): Promise<FolderInfo[]> {
  if (realBindings) return realBindings.GetFolders();
  return mockFolders;
}

export async function GetDisks(): Promise<DiskInfo[]> {
  if (realBindings) return realBindings.GetDisks();
  return mockDisks;
}

export async function CheckOneDriveKFM(): Promise<{ isActive: boolean; folders: string[]; hasGPOBlock: boolean }> {
  if (realBindings) return realBindings.CheckOneDriveKFM();
  return { isActive: false, folders: [], hasGPOBlock: false };
}

export async function StartMigration(targetDrive: string): Promise<void> {
  if (realBindings) return realBindings.StartMigration(targetDrive);
  console.log('Mock: StartMigration to', targetDrive);
}

export async function CancelMigration(): Promise<void> {
  if (realBindings) return realBindings.CancelMigration();
}

export async function Rollback(folderName: string): Promise<void> {
  if (realBindings) return realBindings.Rollback(folderName);
}

export async function HealthCheck(): Promise<boolean> {
  if (realBindings) return realBindings.HealthCheck();
  return true;
}

export async function GetVersion(): Promise<string> {
  if (realBindings) return realBindings.GetVersion();
  return '0.1.0';
}

export async function GetFirstRunFlag(): Promise<boolean> {
  if (realBindings) return realBindings.GetFirstRunFlag();
  return true;
}

export async function SetFirstRunFlag(firstRun: boolean): Promise<void> {
  if (realBindings) return realBindings.SetFirstRunFlag(firstRun);
}

export async function EstimateDuration(targetDrive: string): Promise<any> {
  if (realBindings) return realBindings.EstimateDuration(targetDrive);
  return { isCrossVolume: true, targetIsSSD: true, estimatedSec: 60, estimatedText: '约 1 分钟' };
}

export async function StartupIntegrityCheck(): Promise<any[]> {
  if (realBindings) return realBindings.StartupIntegrityCheck();
  return [];
}

export async function DetectCloudSync(): Promise<any[]> {
  if (realBindings) return realBindings.DetectCloudSync();
  return [];
}

export async function CheckUpdate(): Promise<any> {
  if (realBindings) return realBindings.CheckUpdate();
  return { hasUpdate: false, newVersion: '0.1.0', downloadURL: '' };
}

// ---- Wails Event Subscriptions ----

// Type for the event callback matching Go backend ProgressInfo
type ProgressCallback = (progress: ProgressInfo) => void;

// Type for the migration:done event payload
interface MigrationDoneEvent {
  success: boolean;
  folder?: string;
  error?: string;
  code?: string;
  detail?: string;
  action?: string;
  warning?: string;
}

type MigrationDoneCallback = (event: MigrationDoneEvent) => void;

/**
 * Subscribe to real-time migration progress events from the Go backend.
 * The backend emits `migration:progress` via runtime.EventsEmit in app.go.
 * Returns an unsubscribe function.
 */
export function onMigrationProgress(callback: ProgressCallback): () => void {
  if (realBindings && (window as any).runtime?.EventsOn) {
    return (window as any).runtime.EventsOn('migration:progress', callback);
  }
  // Mock mode: no-op subscription, return empty unsubscribe
  console.warn('onMigrationProgress: Wails runtime not available, using mock');
  return () => {};
}

/**
 * Subscribe to migration completion/failure events from the Go backend.
 * The backend emits `migration:done` via runtime.EventsEmit in app.go.
 * Returns an unsubscribe function.
 */
export function onMigrationDone(callback: MigrationDoneCallback): () => void {
  if (realBindings && (window as any).runtime?.EventsOn) {
    return (window as any).runtime.EventsOn('migration:done', callback);
  }
  console.warn('onMigrationDone: Wails runtime not available, using mock');
  return () => {};
}
