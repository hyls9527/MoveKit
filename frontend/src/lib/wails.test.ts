import { describe, it, expect } from 'vitest'

// Test the Wails binding types match expected shapes
import type { FolderInfo, DiskInfo, ProgressInfo, AppState } from '../types'

describe('Wails type definitions', () => {
  it('FolderInfo has required fields', () => {
    const folder: FolderInfo = {
      name: '测试',
      guid: '{test}',
      sourcePath: 'C:\\test',
      targetPath: 'D:\\test',
      sizeBytes: 1024,
      fileCount: 10,
      status: 'default',
    }
    expect(folder.name).toBe('测试')
    expect(folder.status).toBe('default')
  })

  it('DiskInfo has required fields', () => {
    const disk: DiskInfo = {
      drive: 'D:',
      label: '数据盘',
      totalBytes: 500e9,
      freeBytes: 200e9,
      isFixedDrive: true,
      isRemovable: false,
      isNetwork: false,
      fileSystem: 'NTFS',
    }
    expect(disk.drive).toBe('D:')
    expect(disk.isFixedDrive).toBe(true)
  })

  it('ProgressInfo has required fields', () => {
    const progress: ProgressInfo = {
      folder: '桌面',
      phase: 'moving',
      filesMoved: 42,
      totalFiles: 100,
      bytesMoved: 1000000,
      totalBytes: 5000000,
      percent: 20,
    }
    expect(progress.percent).toBe(20)
  })

  it('AppState type covers all states', () => {
    const states: AppState[] = ['idle', 'checking', 'confirming', 'migrating', 'completed', 'failed']
    expect(states).toHaveLength(6)
  })
})
