import { ProgressInfo } from '../types';

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return (bytes / 1e12).toFixed(1) + ' TB';
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB';
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(1) + ' MB';
  return (bytes / 1e3).toFixed(1) + ' KB';
}

interface ProgressBarProps {
  progress: ProgressInfo;
  onCancel?: () => void;
}

export default function ProgressBar({ progress, onCancel }: ProgressBarProps) {
  const pct = Math.min(100, Math.max(0, progress.percent));

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold text-text-primary">正在迁移</h2>

      {/* Current folder & phase */}
      <div className="text-sm text-text-secondary">
        <span className="font-medium text-text-primary">{progress.folder}</span>
        <span className="mx-2">·</span>
        <span>{progress.phase}</span>
      </div>

      {/* Byte progress bar */}
      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-text-secondary">进度</span>
          <span className="font-medium text-text-primary">{pct.toFixed(1)}%</span>
        </div>
        <div className="w-full h-3 bg-surface-alt rounded-full overflow-hidden">
          <div
            className="h-full bg-accent rounded-full transition-all duration-300 ease-out"
            style={{ width: `${pct}%` }}
          />
        </div>
        <div className="flex justify-between text-xs text-text-secondary">
          <span>{formatBytes(progress.bytesMoved)}</span>
          <span>{formatBytes(progress.totalBytes)}</span>
        </div>
      </div>

      {/* File count progress */}
      <div className="space-y-2">
        <div className="flex justify-between text-sm">
          <span className="text-text-secondary">文件数</span>
          <span className="font-medium text-text-primary">
            {progress.filesMoved} / {progress.totalFiles}
          </span>
        </div>
        <div className="w-full h-2 bg-surface-alt rounded-full overflow-hidden">
          <div
            className="h-full bg-accent-light rounded-full transition-all duration-300 ease-out"
            style={{
              width: `${progress.totalFiles > 0 ? (progress.filesMoved / progress.totalFiles) * 100 : 0}%`,
            }}
          />
        </div>
      </div>

      {/* Warning */}
      <div className="flex items-center gap-2 p-3 rounded-fluent bg-warning/10 border border-warning/20">
        <span className="text-warning">⚠</span>
        <span className="text-sm text-warning">迁移过程中请勿关闭程序或关闭电脑</span>
      </div>

      {/* Cancel button */}
      {onCancel && (
        <button
          onClick={onCancel}
          className="px-4 py-2 text-sm text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
        >
          取消迁移
        </button>
      )}
    </div>
  );
}
