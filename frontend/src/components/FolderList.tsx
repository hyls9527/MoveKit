import { FolderInfo } from '../types';

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return (bytes / 1e12).toFixed(1) + ' TB';
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB';
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(1) + ' MB';
  return (bytes / 1e3).toFixed(1) + ' KB';
}

const statusConfig: Record<string, { label: string; color: string; bg: string }> = {
  default: { label: '待迁移', color: 'text-text-secondary', bg: 'bg-surface-alt' },
  migrated: { label: '已迁移', color: 'text-success', bg: 'bg-success/10' },
  not_on_c: { label: '无需迁移', color: 'text-blue-500', bg: 'bg-blue-500/10' },
  unknown: { label: '未知', color: 'text-text-secondary', bg: 'bg-surface-alt' },
  junction_broken: { label: '链接异常', color: 'text-danger', bg: 'bg-danger/10' },
};

interface FolderListProps {
  folders: FolderInfo[];
}

export default function FolderList({ folders }: FolderListProps) {
  return (
    <div className="space-y-3">
      <h2 className="text-lg font-semibold text-text-primary">系统文件夹</h2>
      <div className="space-y-2">
        {folders.map(folder => {
          const status = statusConfig[folder.status] ?? statusConfig.default;
          return (
            <div
              key={folder.guid}
              className="flex items-center justify-between p-4 rounded-fluent bg-card border border-border"
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-text-primary">{folder.name}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${status.bg} ${status.color}`}>
                    {status.label}
                  </span>
                </div>
                <p className="text-sm text-text-secondary mt-1 truncate">{folder.sourcePath}</p>
              </div>
              <div className="text-right ml-4 flex-shrink-0">
                <span className="text-sm font-medium text-text-primary">{formatBytes(folder.sizeBytes)}</span>
                <p className="text-xs text-text-secondary">{folder.fileCount} 个文件</p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
