import { DiskInfo } from '../types';

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return (bytes / 1e12).toFixed(1) + ' TB';
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB';
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(1) + ' MB';
  return (bytes / 1e3).toFixed(1) + ' KB';
}

interface DiskSelectorProps {
  disks: DiskInfo[];
  selected: string;
  onSelect: (drive: string) => void;
}

export default function DiskSelector({ disks, selected, onSelect }: DiskSelectorProps) {
  // Recommend the disk with the most free space (only eligible ones)
  const eligible = disks.filter(
    d => !d.isRemovable && !d.isNetwork && d.fileSystem === 'NTFS' && d.drive.toUpperCase() !== 'C:',
  );
  const recommended = eligible.reduce<DiskInfo | null>(
    (best, d) => (!best || d.freeBytes > best.freeBytes ? d : best),
    null,
  );

  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold text-text-primary">选择目标磁盘</h2>
      <div className="space-y-2" role="radiogroup" aria-label="选择目标磁盘">
        {disks.map(disk => {
          const disabled =
            disk.isRemovable ||
            disk.isNetwork ||
            disk.fileSystem !== 'NTFS' ||
            disk.drive.toUpperCase() === 'C:';
          const isRecommended = recommended?.drive === disk.drive;
          const isSelected = selected === disk.drive;

          return (
            <label
              key={disk.drive}
              role="radio"
              aria-checked={isSelected}
              aria-disabled={disabled}
              className={`
                flex items-center gap-4 p-4 rounded-fluent border transition-shadow
                ${disabled
                  ? 'bg-surface-alt border-border text-text-disabled cursor-not-allowed opacity-60'
                  : isSelected
                    ? 'bg-card border-accent shadow-fluent cursor-pointer ring-2 ring-accent/20'
                    : 'bg-card border-border hover:border-border-hover hover:shadow-fluent cursor-pointer'
                }
              `}
            >
              <input
                type="radio"
                name="disk"
                value={disk.drive}
                disabled={disabled}
                checked={isSelected}
                onChange={() => onSelect(disk.drive)}
                className="sr-only"
              />
              <div className={`
                w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0
                ${isSelected ? 'border-accent' : 'border-text-disabled'}
              `}>
                {isSelected && <div className="w-2.5 h-2.5 rounded-full bg-accent" />}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-semibold text-text-primary">{disk.drive}</span>
                  <span className="text-text-secondary text-sm">{disk.label}</span>
                  {/* SSD/HDD type badge removed: reliable detection not yet implemented */}
                  {isRecommended && (
                    <span className="text-xs px-2 py-0.5 rounded-full bg-accent/10 text-accent font-medium">
                      推荐
                    </span>
                  )}
                </div>
                <div className="text-sm text-text-secondary mt-1">
                  {formatBytes(disk.freeBytes)} 可用 / {formatBytes(disk.totalBytes)} 总计
                  {disk.fileSystem !== 'NTFS' && <span className="ml-2 text-danger">（{disk.fileSystem} 不支持）</span>}
                  {disk.isRemovable && <span className="ml-2 text-warning">（可移动磁盘）</span>}
                  {disk.isNetwork && <span className="ml-2 text-warning">（网络磁盘）</span>}
                  {disk.drive.toUpperCase() === 'C:' && (
                    <span className="ml-2 text-warning">（源盘不可作为迁移目标）</span>
                  )}
                </div>
              </div>
            </label>
          );
        })}
      </div>

      {!selected && (
        <p className="text-sm text-text-secondary">请选择一个目标磁盘以继续</p>
      )}
      {/* HDD warning removed: SSD vs HDD detection not yet implemented */}
    </div>
  );
}
