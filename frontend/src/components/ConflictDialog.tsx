export interface ConflictFile {
  filename: string;
  sourceSize: number;
  sourceDate: string;
  targetSize: number;
  targetDate: string;
}

interface ConflictDialogProps {
  conflicts: ConflictFile[];
  onSkip: () => void;
  onOverwrite: () => void;
  onCancel: () => void;
}

function formatBytes(bytes: number): string {
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB';
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(1) + ' MB';
  if (bytes >= 1e3) return (bytes / 1e3).toFixed(1) + ' KB';
  return bytes + ' B';
}

export default function ConflictDialog({ conflicts, onSkip, onOverwrite, onCancel }: ConflictDialogProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-2xl w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-3">
            <span className="text-2xl">⚠️</span>
            <h2 className="text-lg font-semibold text-text-primary">
              检测到文件名冲突
            </h2>
          </div>
          <p className="text-sm text-text-secondary mt-2">
            目标路径中已存在以下 {conflicts.length} 个同名文件，请选择处理方式。
          </p>
        </div>

        {/* Conflict table */}
        <div className="px-6 py-4 max-h-64 overflow-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-text-secondary text-xs">
                <th className="text-left py-2 pr-3 font-medium">文件名</th>
                <th className="text-right py-2 px-3 font-medium">源文件大小</th>
                <th className="text-right py-2 px-3 font-medium">源文件日期</th>
                <th className="text-right py-2 px-3 font-medium">目标大小</th>
                <th className="text-right py-2 pl-3 font-medium">目标日期</th>
              </tr>
            </thead>
            <tbody>
              {conflicts.map((c) => (
                <tr key={c.filename} className="border-b border-border/50 last:border-0">
                  <td className="py-2 pr-3 text-text-primary font-medium truncate max-w-[200px]">
                    {c.filename}
                  </td>
                  <td className="py-2 px-3 text-right text-text-secondary">
                    {formatBytes(c.sourceSize)}
                  </td>
                  <td className="py-2 px-3 text-right text-text-secondary">
                    {c.sourceDate}
                  </td>
                  <td className="py-2 px-3 text-right text-text-secondary">
                    {formatBytes(c.targetSize)}
                  </td>
                  <td className="py-2 pl-3 text-right text-text-secondary">
                    {c.targetDate}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-border flex justify-end gap-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
          >
            取消迁移
          </button>
          <button
            onClick={onSkip}
            className="px-4 py-2 text-sm font-medium text-accent rounded-fluent border border-accent/30 hover:bg-accent/5 transition-colors"
          >
            跳过冲突文件
          </button>
          <button
            onClick={onOverwrite}
            className="px-4 py-2 text-sm font-medium text-white rounded-fluent bg-warning hover:bg-warning/90 transition-colors"
          >
            覆盖目标文件
          </button>
        </div>
      </div>
    </div>
  );
}
