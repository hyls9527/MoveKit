import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import DiskSelector from './DiskSelector'
import type { DiskInfo } from '../types'

const disks: DiskInfo[] = [
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

describe('DiskSelector', () => {
  it('disables the source drive C: so it cannot be selected as target', () => {
    const { container } = render(<DiskSelector disks={disks} selected="" onSelect={() => {}} />)
    const radios = Array.from(container.querySelectorAll('input[type="radio"]')) as HTMLInputElement[]
    const c = radios.find(r => r.value === 'C:')
    const d = radios.find(r => r.value === 'D:')
    expect(c).toBeDefined()
    expect(d).toBeDefined()
    expect(c!.disabled).toBe(true)
    expect(d!.disabled).toBe(false)
    expect(screen.getByText('（源盘不可作为迁移目标）')).toBeDefined()
  })

  it('recommends the eligible disk with the most free space (never C:)', () => {
    render(<DiskSelector disks={disks} selected="" onSelect={() => {}} />)
    const badges = screen.getAllByText('推荐')
    expect(badges).toHaveLength(1)
    const recommendedRow = badges[0].closest('label')
    expect(recommendedRow?.textContent).toContain('D:')
    expect(recommendedRow?.textContent).not.toContain('C:')
  })

  it('fires onSelect when an enabled drive is chosen', () => {
    const onSelect = vi.fn()
    const { container } = render(<DiskSelector disks={disks} selected="" onSelect={onSelect} />)
    const radios = Array.from(container.querySelectorAll('input[type="radio"]')) as HTMLInputElement[]
    fireEvent.click(radios.find(r => r.value === 'D:')!)
    expect(onSelect).toHaveBeenCalledWith('D:')
  })
})
