interface OneDriveWarningProps {
  folders: string[];
  onClose: () => void;
}

export default function OneDriveWarning({ folders, onClose }: OneDriveWarningProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-md w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-3">
            <span className="text-2xl">⚠️</span>
            <h2 className="text-lg font-semibold text-text-primary">
              检测到 OneDrive 已接管您的文件夹
            </h2>
          </div>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4">
          <p className="text-sm text-text-secondary">
            OneDrive 的「已知文件夹迁移（KFM）」功能正在将以下文件夹同步到云端。
            在此状态下进行迁移可能导致数据丢失或同步冲突。
          </p>

          {/* Affected folders */}
          <div className="p-3 rounded-fluent bg-warning/5 border border-warning/20">
            <p className="text-xs font-medium text-warning mb-2">受影响的文件夹：</p>
            <ul className="space-y-1">
              {folders.map((f) => (
                <li key={f} className="text-sm text-text-primary flex items-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-warning" />
                  {f}
                </li>
              ))}
            </ul>
          </div>

          {/* Steps to disable */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-text-secondary">请按以下步骤关闭 OneDrive 文件夹接管：</p>
            <ol className="space-y-2 text-sm text-text-secondary">
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">1</span>
                <span>右键点击任务栏中的 OneDrive 图标，选择「设置」</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">2</span>
                <span>进入「备份」选项卡</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">3</span>
                <span>点击「管理备份」，关闭所有已接管的文件夹</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">4</span>
                <span>等待 OneDrive 停止同步后，重新打开 MoveKit</span>
              </li>
            </ol>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-border flex justify-end">
          <button
            onClick={onClose}
            className="px-5 py-2 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
          >
            我知道了
          </button>
        </div>
      </div>
    </div>
  );
}
