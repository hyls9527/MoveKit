import { useState } from 'react';

interface FirstRunGuideProps {
  onComplete: () => void;
}

const steps = [
  {
    title: '释放 C 盘空间',
    description: 'MoveKit 帮助您把 C 盘的文件夹搬到其他盘，释放宝贵的系统盘空间。',
    illustration: (
      <div className="flex items-center justify-center gap-6 py-6">
        <div className="flex flex-col items-center gap-2">
          <div className="w-16 h-16 rounded-fluent bg-danger/10 border border-danger/20 flex items-center justify-center text-2xl">
            💻
          </div>
          <span className="text-xs text-text-secondary">C 盘（空间不足）</span>
        </div>
        <div className="text-2xl text-accent animate-pulse">→</div>
        <div className="flex flex-col items-center gap-2">
          <div className="w-16 h-16 rounded-fluent bg-success/10 border border-success/20 flex items-center justify-center text-2xl">
            💾
          </div>
          <span className="text-xs text-text-secondary">其他盘（充足空间）</span>
        </div>
      </div>
    ),
  },
  {
    title: '自动迁移',
    description: '选择目标磁盘，MoveKit 会自动迁移所有文件，并创建文件夹链接确保程序正常运行。',
    illustration: (
      <div className="flex flex-col items-center gap-3 py-6">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-accent/10 flex items-center justify-center text-accent">
            📁
          </div>
          <div className="w-10 h-10 rounded-full bg-accent/10 flex items-center justify-center text-accent">
            📷
          </div>
          <div className="w-10 h-10 rounded-full bg-accent/10 flex items-center justify-center text-accent">
            📄
          </div>
          <div className="w-10 h-10 rounded-full bg-accent/10 flex items-center justify-center text-accent">
            🎵
          </div>
        </div>
        <svg className="w-5 h-5 text-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
        </svg>
        <div className="px-4 py-2 rounded-fluent bg-accent/10 border border-accent/20">
          <span className="text-sm text-accent font-medium">一键迁移完成 ✓</span>
        </div>
      </div>
    ),
  },
  {
    title: '安全可靠',
    description: '迁移是安全的，所有操作都可以随时回滚。您的数据不会丢失。',
    illustration: (
      <div className="flex flex-col items-center gap-3 py-6">
        <div className="w-16 h-16 rounded-full bg-success/10 border-2 border-success/30 flex items-center justify-center">
          <span className="text-3xl">🔒</span>
        </div>
        <div className="flex gap-4 text-xs text-text-secondary">
          <span className="flex items-center gap-1">
            <span className="text-success">✓</span> 可回滚
          </span>
          <span className="flex items-center gap-1">
            <span className="text-success">✓</span> 数据安全
          </span>
          <span className="flex items-center gap-1">
            <span className="text-success">✓</span> 不影响程序
          </span>
        </div>
      </div>
    ),
  },
];

export default function FirstRunGuide({ onComplete }: FirstRunGuideProps) {
  const [current, setCurrent] = useState(0);
  const isLast = current === steps.length - 1;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-md w-full mx-4 overflow-hidden">
        {/* Step indicator */}
        <div className="flex justify-center gap-2 pt-6">
          {steps.map((_, i) => (
            <div
              key={i}
              className={`h-1.5 rounded-full transition-all duration-300 ${
                i === current ? 'w-8 bg-accent' : i < current ? 'w-4 bg-accent/40' : 'w-4 bg-surface-alt'
              }`}
            />
          ))}
        </div>

        {/* Illustration */}
        <div className="px-6">
          {steps[current].illustration}
        </div>

        {/* Content */}
        <div className="px-6 pb-2 text-center space-y-2">
          <h2 className="text-xl font-semibold text-text-primary">
            {steps[current].title}
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed">
            {steps[current].description}
          </p>
        </div>

        {/* Footer */}
        <div className="px-6 py-5 flex justify-between items-center">
          {!isLast ? (
            <button
              onClick={onComplete}
              className="text-sm text-text-secondary hover:text-text-primary transition-colors"
            >
              跳过
            </button>
          ) : (
            <div />
          )}

          {!isLast ? (
            <button
              onClick={() => setCurrent((c) => c + 1)}
              className="px-6 py-2.5 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
            >
              下一步
            </button>
          ) : (
            <button
              onClick={onComplete}
              className="px-6 py-2.5 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
            >
              开始使用
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
