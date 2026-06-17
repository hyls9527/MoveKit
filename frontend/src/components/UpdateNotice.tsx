interface UpdateNoticeProps {
  currentVersion: string;
  newVersion: string;
  downloadUrl: string;
  onLater: () => void;
}

export default function UpdateNotice({ currentVersion, newVersion, downloadUrl, onLater }: UpdateNoticeProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-sm w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🔄</span>
            <h2 className="text-lg font-semibold text-text-primary">
              发现新版本
            </h2>
          </div>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4">
          <p className="text-sm text-text-secondary">
            MoveKit 有新的版本可用，建议更新以获得最新功能和修复。
          </p>

          {/* Version comparison */}
          <div className="flex items-center gap-4 p-4 rounded-fluent bg-surface-alt">
            <div className="flex-1 text-center">
              <p className="text-xs text-text-secondary mb-1">当前版本</p>
              <p className="text-sm font-semibold text-text-primary">{currentVersion}</p>
            </div>
            <div className="text-text-disabled">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
              </svg>
            </div>
            <div className="flex-1 text-center">
              <p className="text-xs text-text-secondary mb-1">最新版本</p>
              <p className="text-sm font-semibold text-accent">{newVersion}</p>
            </div>
          </div>

          <p className="text-xs text-text-secondary">
            点击「下载更新」将打开浏览器下载最新版本安装包。
          </p>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-border flex justify-end gap-3">
          <button
            onClick={onLater}
            className="px-4 py-2 text-sm text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
          >
            稍后
          </button>
          <a
            href={downloadUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="px-5 py-2 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors inline-block"
          >
            下载更新
          </a>
        </div>
      </div>
    </div>
  );
}
