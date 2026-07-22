interface DefenderInfoProps {
  onClose: () => void;
}

export default function DefenderInfo({ onClose }: DefenderInfoProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-lg w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="px-6 pt-6 pb-2">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🛡️</span>
            <h2 className="text-lg font-semibold text-text-primary">
              关于 Windows 安全提示
            </h2>
          </div>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-5">
          <p className="text-sm text-text-secondary">
            MoveKit 需要使用管理员权限创建文件夹链接（Junction），这可能触发
            Windows SmartScreen 或 Defender 的误报提示。MoveKit 是开源软件，
            不含任何恶意代码，您可以安全使用。
          </p>

          {/* SmartScreen section */}
          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-text-primary flex items-center gap-2">
              <span className="text-accent">🔷</span>
              Windows SmartScreen 提示
            </h3>
            <p className="text-xs text-text-secondary">如果您看到「Windows 已保护您的电脑」的蓝色提示：</p>
            <ol className="space-y-1.5 text-sm text-text-secondary ml-1">
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">1</span>
                <span>点击「更多信息」</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">2</span>
                <span>确认发布者为「MoveKit」后，点击「仍要运行」</span>
              </li>
            </ol>
          </div>

          {/* Defender section */}
          <div className="space-y-2">
            <h3 className="text-sm font-semibold text-text-primary flex items-center gap-2">
              <span className="text-danger">🛡️</span>
              Windows Defender 误报
            </h3>
            <p className="text-xs text-text-secondary">如果 Defender 拦截了 MoveKit：</p>
            <ol className="space-y-1.5 text-sm text-text-secondary ml-1">
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">1</span>
                <span>打开「Windows 安全中心」→「病毒和威胁防护」</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">2</span>
                <span>在「保护历史记录」中找到被拦截的 MoveKit</span>
              </li>
              <li className="flex gap-2">
                <span className="flex-shrink-0 w-5 h-5 rounded-full bg-accent/10 text-accent text-xs flex items-center justify-center font-medium">3</span>
                <span>点击「操作」→「允许」</span>
              </li>
            </ol>
          </div>

          <div className="p-3 rounded-fluent bg-accent/5 border border-accent/10">
            <p className="text-xs text-accent">
              💡 MoveKit 已开源，您可以在 GitHub 上查看完整源代码以确认安全性。
            </p>
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
