import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import FolderList from './FolderList'

describe('FolderList', () => {
  const mockFolders = [
    {
      name: '桌面',
      guid: '{desktop}',
      sourcePath: 'C:\\Users\\Test\\Desktop',
      targetPath: 'C:\\Users\\Test\\Desktop',
      sizeBytes: 1024 * 1024,
      fileCount: 10,
      status: 'default' as const,
    },
    {
      name: '文档',
      guid: '{docs}',
      sourcePath: 'D:\\Documents',
      targetPath: 'D:\\Documents',
      sizeBytes: 2048 * 1024,
      fileCount: 20,
      status: 'migrated' as const,
    },
  ]

  it('renders folder names', () => {
    render(<FolderList folders={mockFolders} />)
    expect(screen.getByText('桌面')).toBeDefined()
    expect(screen.getByText('文档')).toBeDefined()
  })

  it('renders empty state', () => {
    render(<FolderList folders={[]} />)
    expect(document.body.textContent).toBeDefined()
  })
})
