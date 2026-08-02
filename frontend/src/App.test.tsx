import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'

interface MigrationDoneEvent {
  success: boolean
  error?: string
  folder?: string
  code?: string
  detail?: string
  action?: string
  warning?: string
}

const mocks = vi.hoisted(() => {
  const doneCallbacks: Array<(e: MigrationDoneEvent) => void> = []
  return {
    doneCallbacks,
    getFolders: vi.fn(),
    getDisks: vi.fn(),
    startMigration: vi.fn(),
    checkOneDrive: vi.fn(),
    cancelMigration: vi.fn(),
    onProgress: vi.fn(),
    onDone: vi.fn(),
    healthCheck: vi.fn(),
  }
})

vi.mock('./lib/wails', () => ({
  GetFolders: mocks.getFolders,
  GetDisks: mocks.getDisks,
  StartMigration: mocks.startMigration,
  Rollback: vi.fn(),
  CheckOneDriveKFM: mocks.checkOneDrive,
  CancelMigration: mocks.cancelMigration,
  onMigrationProgress: mocks.onProgress,
  onMigrationDone: mocks.onDone,
  HealthCheck: mocks.healthCheck,
}))

import App from './App'

const folders = [
  {
    name: '桌面',
    guid: '{desktop}',
    sourcePath: 'C:\\Users\\Test\\Desktop',
    targetPath: '',
    sizeBytes: 2048,
    fileCount: 1,
    status: 'default',
  },
  {
    name: '文档',
    guid: '{docs}',
    sourcePath: 'D:\\Users\\Test\\Documents',
    targetPath: 'D:\\Users\\Test\\Documents',
    sizeBytes: 5000,
    fileCount: 1,
    status: 'migrated',
  },
]

const disks = [
  {
    drive: 'C:',
    label: '系统盘',
    totalBytes: 256e9,
    freeBytes: 45e9,
    isFixedDrive: true,
    isRemovable: false,
    isNetwork: false,
    fileSystem: 'NTFS',
  },
  {
    drive: 'D:',
    label: '数据盘',
    totalBytes: 512e9,
    freeBytes: 128e9,
    isFixedDrive: true,
    isRemovable: false,
    isNetwork: false,
    fileSystem: 'NTFS',
  },
]

function setup() {
  localStorage.setItem('movekit-first-run-seen', 'true')
  mocks.doneCallbacks.length = 0
  mocks.getFolders.mockReset().mockResolvedValue(folders)
  mocks.getDisks.mockReset().mockResolvedValue(disks)
  mocks.startMigration.mockReset().mockResolvedValue(undefined)
  mocks.checkOneDrive.mockReset().mockResolvedValue({ isActive: false, folders: [], hasGPOBlock: false })
  mocks.cancelMigration.mockReset().mockResolvedValue(undefined)
  mocks.onProgress.mockReset().mockReturnValue(() => {})
  mocks.onDone.mockReset().mockImplementation((cb: (e: MigrationDoneEvent) => void) => {
    mocks.doneCallbacks.push(cb)
    return () => {}
  })
  mocks.healthCheck.mockReset().mockResolvedValue(true)
}

async function startMigrationToConfirmPage() {
  const { container } = render(<App />)
  await screen.findByText('桌面')
  const radios = Array.from(container.querySelectorAll('input[type="radio"]')) as HTMLInputElement[]
  fireEvent.click(radios.find(r => r.value === 'D:')!)
  await waitFor(() => expect(screen.getByRole('button', { name: '开始迁移' })).not.toBeDisabled())
  fireEvent.click(screen.getByRole('button', { name: '开始迁移' }))
  await waitFor(() => expect(mocks.checkOneDrive).toHaveBeenCalled())
  await screen.findByRole('button', { name: '确认迁移' })
  fireEvent.click(screen.getByRole('button', { name: '确认迁移' }))
  await waitFor(() => expect(mocks.startMigration).toHaveBeenCalledWith('D:'))
  await waitFor(() => expect(mocks.onDone).toHaveBeenCalled())
}

describe('App migration flow', () => {
  beforeEach(() => {
    setup()
  })

  it('shows completed page and refreshes folder state after success', async () => {
    await startMigrationToConfirmPage()
    const doneCb = mocks.doneCallbacks[0]
    doneCb({ success: true })
    expect(await screen.findByText('迁移完成')).toBeDefined()
    expect(mocks.getFolders).toHaveBeenCalledTimes(2)
    // Freed space must count only folders migrated in THIS run (桌面 2048),
    // not the folder that was already migrated (文档 5000).
    expect(screen.getAllByText('2.0 KB').length).toBeGreaterThanOrEqual(1)
    expect(screen.queryByText('7.0 KB')).toBeNull()
  })

  it('shows failure message and still refreshes folders for rollback', async () => {
    await startMigrationToConfirmPage()
    const doneCb = mocks.doneCallbacks[0]
    doneCb({ success: false, error: 'boom', action: '请重试' })
    expect(await screen.findByText('迁移失败：boom（请重试）')).toBeDefined()
    expect(mocks.getFolders).toHaveBeenCalledTimes(2)
  })

  it('shows cancel message and returns to idle when migration is cancelled', async () => {
    await startMigrationToConfirmPage()
    const doneCb = mocks.doneCallbacks[0]
    doneCb({ success: false, code: 'CANCELLED', error: '迁移已取消，正在恢复已移动的文件' })
    expect(await screen.findByText('迁移已取消，已移动的文件已恢复到原位置')).toBeDefined()
    expect(screen.getByRole('button', { name: '开始迁移' })).toBeDefined()
  })
})
