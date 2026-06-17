interface CloudWarningProps {
  providers: string[];
  onClose: () => void;
}

const providerIcons: Record<string, string> = {
  OneDrive: '☁️',
  'Google Drive': '🔷',
  Dropbox: '💧',
  iCloud: '🍎',
  '坚果云': '🌰',
  百度网盘: '💾',
};

export default function CloudWarning({ providers, onClose }: CloudWarningProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-md w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-3">
            <span className="text-2xl">☁️</span>
            <h2 className="text-lg font-semibold text-text-primary">
              检测到云同步服务正在运行
            </h2>
          </div>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4">
          <p className="text-sm text-text-secondary">
            以下云同步服务正在后台运行。在迁移过程中，云同步可能会干扰文件移动操作，
            导致文件冲突或数据丢失。建议先关闭云同步服务再继续。
          </p>

          {/* Detected providers */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-text-secondary">已检测到的服务：</p>
            <div className="p-3 rounded-fluent bg-warning/5 border border-warning/20">
              <ul className="space-y-2">
                {providers.map((p) => (
                  <li key={p} className="flex items-center gap-3 text-sm text-text-primary">
                    <span>{providerIcons[p] ?? '☁️'}</span>
                    <span className="font-medium">{p}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>

          {/* Suggestions */}
          <div className="space-y-2">
            <p className="text-xs font-medium text-text-secondary">建议操作：</p>
            <ol className="space-y-1.5 text-sm text-text-secondary">
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">1</span>
                <span>右键点击系统托盘中的云同步图标</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">2</span>
                <span>选择「暂停同步」或「退出」</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">3</span>
                <span>完成迁移后再重新启动同步服务</span>
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
