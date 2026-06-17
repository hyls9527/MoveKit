import { useState } from 'react';

interface HelpPanelProps {
  open: boolean;
  onClose: () => void;
}

const faqs = [
  {
    question: '什么是 Junction？',
    answer:
      'Junction（联接点）是 Windows 提供的一种特殊文件夹链接。MoveKit 使用 Junction 将原始路径指向新位置，这样所有程序仍然能通过原路径访问文件，您无需修改任何设置。',
  },
  {
    question: '迁移后文件在哪里？',
    answer:
      '文件被完整移动到您选择的目标磁盘。例如，如果您将桌面迁移到 D 盘，文件实际位于 D:\\Users\\用户名\\Desktop，但 C 盘的原路径通过 Junction 仍然可用。',
  },
  {
    question: '如何回滚？',
    answer:
      '在 MoveKit 的迁移结果页面，每个已迁移的文件夹旁边都有「回滚」按钮。点击后，MoveKit 会删除 Junction 并将文件移回原位置。回滚不会丢失任何数据。',
  },
  {
    question: '迁移会影响其他程序吗？',
    answer:
      '正常使用不会受影响。Junction 机制确保所有程序（包括 Office、浏览器等）通过原路径依然能找到文件。极少数老旧程序可能不支持 Junction，如遇到问题可以回滚。',
  },
];

export default function HelpPanel({ open, onClose }: HelpPanelProps) {
  const [expanded, setExpanded] = useState<number | null>(null);

  return (
    <>
      {/* Backdrop */}
      {open && (
        <div
          className="fixed inset-0 z-40 bg-black/20 transition-opacity"
          onClick={onClose}
        />
      )}

      {/* Panel */}
      <div
        className={`fixed top-0 right-0 z-50 h-full w-96 max-w-[90vw] bg-card border-l border-border shadow-fluent-lg transform transition-transform duration-300 ${
          open ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold text-text-primary">帮助与反馈</h2>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-fluent flex items-center justify-center text-text-secondary hover:bg-surface-alt transition-colors"
            aria-label="关闭"
          >
            ✕
          </button>
        </div>

        {/* FAQ */}
        <div className="px-6 py-4 space-y-3 overflow-y-auto" style={{ height: 'calc(100% - 120px)' }}>
          <h3 className="text-sm font-medium text-text-secondary mb-3">常见问题</h3>

          {faqs.map((faq, i) => (
            <div key={i} className="border border-border rounded-fluent overflow-hidden">
              <button
                onClick={() => setExpanded(expanded === i ? null : i)}
                className="w-full px-4 py-3 text-left text-sm font-medium text-text-primary flex items-center justify-between hover:bg-surface-alt transition-colors"
              >
                <span>{faq.question}</span>
                <svg
                  className={`w-4 h-4 text-text-secondary transition-transform duration-200 ${
                    expanded === i ? 'rotate-180' : ''
                  }`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {expanded === i && (
                <div className="px-4 pb-3 text-sm text-text-secondary leading-relaxed border-t border-border/50 pt-2">
                  {faq.answer}
                </div>
              )}
            </div>
          ))}

          {/* Contact / feedback */}
          <div className="mt-6 pt-4 border-t border-border">
            <h3 className="text-sm font-medium text-text-secondary mb-3">反馈与支持</h3>
            <div className="space-y-3">
              <a
                href="https://github.com/nousresearch/movekit/issues"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-3 p-3 rounded-fluent border border-border hover:border-accent/30 hover:bg-accent/5 transition-colors"
              >
                <span className="text-lg">🐛</span>
                <div>
                  <p className="text-sm font-medium text-text-primary">提交问题反馈</p>
                  <p className="text-xs text-text-secondary">在 GitHub 上报告 Bug 或建议功能</p>
                </div>
              </a>
              <a
                href="https://github.com/nousresearch/movekit"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-3 p-3 rounded-fluent border border-border hover:border-accent/30 hover:bg-accent/5 transition-colors"
              >
                <span className="text-lg">⭐</span>
                <div>
                  <p className="text-sm font-medium text-text-primary">GitHub 主页</p>
                  <p className="text-xs text-text-secondary">查看源代码、文档与更新日志</p>
                </div>
              </a>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
