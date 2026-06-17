import { FolderInfo } from '../types';

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return (bytes / 1e12).toFixed(1) + ' TB';
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB';
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(1) + ' MB';
  return (bytes / 1e3).toFixed(1) + ' KB';
}

interface ResultPageProps {
  success: boolean;
  folders: FolderInfo[];
  freedBytes: number;
  onRollback: (folderName: string) => void;
  onDone: () => void;
}

export default function ResultPage({ success, folders, freedBytes, onRollback, onDone }: ResultPageProps) {
  const migratedFolders = folders.filter(f => f.status === 'migrated');

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        {success ? (
          <div className="w-10 h-10 rounded-full bg-success/10 flex items-center justify-center">
            <span className="text-success text-xl">✓</span>
          </div>
        ) : (
          <div className="w-10 h-10 rounded-full bg-danger/10 flex items-center justify-center">
            <span className="text-danger text-xl">✕</span>
          </div>
        )}
        <h2 className="text-lg font-semibold text-text-primary">
          {success ? '迁移完成' : '迁移失败'}
        </h2>
      </div>

      {/* Summary */}
      {success && (
        <div className="p-4 rounded-fluent bg-card border border-border">
          <p className="text-text-secondary text-sm">共释放 C 盘空间</p>
          <p className="text-2xl font-bold text-accent">{formatBytes(freedBytes)}</p>
        </div>
      )}

      {/* Folder list */}
      <div className="space-y-2">
        {folders.map(folder => {
          const isMigrated = folder.status === 'migrated';
          return (
            <div
              key={folder.guid}
              className="flex items-center justify-between p-4 rounded-fluent bg-card border border-border"
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  {isMigrated ? (
                    <span className="text-success">✓</span>
                  ) : (
                    <span className="text-danger">✕</span>
                  )}
                  <span className="font-medium text-text-primary">{folder.name}</span>
                </div>
                <p className="text-sm text-text-secondary mt-1 truncate ml-6">
                  {isMigrated && folder.targetPath ? folder.targetPath : folder.sourcePath}
                </p>
              </div>
              {isMigrated && (
                <button
                  onClick={() => onRollback(folder.name)}
                  className="ml-4 px-3 py-1.5 text-xs text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors flex-shrink-0"
                >
                  回滚
                </button>
              )}
            </div>
          );
        })}
      </div>

      {/* Info text */}
      {success && (
        <div className="p-4 rounded-fluent bg-accent/5 border border-accent/10 text-sm text-text-secondary space-y-1">
          <p>• 已通过 NTFS Junction 链接重定向原路径，所有程序和快捷方式无需修改即可正常使用。</p>
          <p>• 迁移前的原始数据将在目标磁盘保留 30 天，届时自动清理。</p>
          <p>• 如遇问题，可随时使用"一键回滚"功能恢复。</p>
        </div>
      )}

      {/* Buttons */}
      <div className="flex gap-3">
        {success && migratedFolders.length > 0 && (
          <button
            onClick={() => {
              migratedFolders.forEach(f => onRollback(f.name));
            }}
            className="px-5 py-2.5 text-sm font-medium text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
          >
            一键回滚
          </button>
        )}
        <button
          onClick={onDone}
          className="px-5 py-2.5 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
        >
          完成
        </button>
      </div>
    </div>
  );
}
