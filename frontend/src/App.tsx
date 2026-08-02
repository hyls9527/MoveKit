import { useState, useEffect, useCallback, useRef } from 'react';
import { AppState, DiskInfo, FolderInfo, ProgressInfo } from './types';
import { GetFolders, GetDisks, StartMigration, Rollback, CheckOneDriveKFM, CancelMigration, onMigrationProgress, onMigrationDone } from './lib/wails';
import DiskSelector from './components/DiskSelector';
import { ErrorBoundary as SectionErrorBoundary } from './components/ErrorBoundary';
import FolderList from './components/FolderList';
import ProgressBar from './components/ProgressBar';
import ResultPage from './components/ResultPage';
import OneDriveWarning from './components/OneDriveWarning';
import DefenderInfo from './components/DefenderInfo';
import { ErrorBoundary, useHeartbeat } from './components/ErrorBoundary';
import FirstRunGuide from './components/FirstRunGuide';
import HelpPanel from './components/HelpPanel';

// The inner app content, wrapped by ErrorBoundary below
function AppInner() {
  const [state, setState] = useState<AppState>('idle');
  const [folders, setFolders] = useState<FolderInfo[]>([]);
  const [disks, setDisks] = useState<DiskInfo[]>([]);
  const [selectedDisk, setSelectedDisk] = useState('');
  const [progress, setProgress] = useState<ProgressInfo>({
    folder: '',
    phase: '准备中',
    filesMoved: 0,
    totalFiles: 0,
    bytesMoved: 0,
    totalBytes: 0,
    percent: 0,
  });
  const [error, setError] = useState('');
  const [freedBytes, setFreedBytes] = useState(0);

  // Dialog visibility states
  const [showOneDriveWarning, setShowOneDriveWarning] = useState(false);
  const [oneDriveFolders, setOneDriveFolders] = useState<string[]>([]);
  const [showDefenderInfo, setShowDefenderInfo] = useState(false);
  const [showFirstRunGuide, setShowFirstRunGuide] = useState(false);
  const [showHelpPanel, setShowHelpPanel] = useState(false);

  // Check first-run on mount
  useEffect(() => {
    const hasSeenGuide = localStorage.getItem('movekit-first-run-seen');
    if (!hasSeenGuide) {
      setShowFirstRunGuide(true);
    }
  }, []);

  // Heartbeat during migration — uses the useHeartbeat hook from ErrorBoundary
  const handleDead = useCallback(() => {
    setError('后端进程失去响应');
    setState('failed');
  }, []);
  useHeartbeat({ enabled: state === 'migrating', onDead: handleDead });

  const loadData = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setState('checking');
    try {
      const [f, d] = await Promise.all([GetFolders(), GetDisks()]);
      setFolders(f);
      setDisks(d);
      if (!opts?.silent) setState('idle');
    } catch {
      if (!opts?.silent) {
        setError('无法加载数据，请检查后端服务');
        setState('failed');
      }
    }
  }, []);

  // Initial load
  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleStart = async () => {
    if (!selectedDisk) return;

    // Check OneDrive KFM before proceeding
    try {
      const oneDrive = await CheckOneDriveKFM();
      if (oneDrive.isActive) {
        setOneDriveFolders(oneDrive.folders);
        setShowOneDriveWarning(true);
        return;
      }
    } catch {
      // If check fails, proceed anyway
    }

    setState('confirming');
  };

  const handleConfirm = async () => {
    preMigrationFoldersRef.current = folders;
    setState('migrating');
    try {
      await StartMigration(selectedDisk);
    } catch (err) {
      setError(`迁移失败：${err instanceof Error ? err.message : '未知错误'}`);
      setState('failed');
    }
  };

  const handleCancel = async () => {
    const confirmed = window.confirm(
      '确定要取消迁移吗？\n' +
      '取消后，已移动的文件会自动恢复到原位置，请等待恢复完成。\n' +
      '跨卷迁移的取消和恢复可能需要一些时间。\n'
    );
    if (!confirmed) return;

    try {
      await CancelMigration();
    } catch {
      // CancelMigration is fire-and-forget; ignore errors
    }
    // The migration:done event will update state to idle/failed.
    // Do NOT set state here — let the backend event drive state transitions.
  };

  // Ref to access latest progress in event callbacks without re-subscribing.
  const progressRef = useRef(progress);
  progressRef.current = progress;
  // Snapshot of folder state when this migration run started, used to report
  // how much space THIS run freed (excluding folders migrated in earlier runs).
  const preMigrationFoldersRef = useRef<FolderInfo[]>([]);

  // ── Wails Event Subscriptions (real progress + completion) ──
  useEffect(() => {
    if (state !== 'migrating') return;

    const unsubProgress = onMigrationProgress((info) => {
      setProgress(info);
    });

    const unsubDone = onMigrationDone((event) => {
      if (event.success) {
        (async () => {
          try {
            const [f, d] = await Promise.all([GetFolders(), GetDisks()]);
            setFolders(f);
            setDisks(d);
            setFreedBytes(
              preMigrationFoldersRef.current
                .filter(fi => fi.status !== 'migrated' && fi.sizeBytes > 0)
                .reduce((sum, fi) => sum + fi.sizeBytes, 0),
            );
          } finally {
            setState('completed');
          }
        })();
      } else {
        if (event.code === 'CANCELLED') {
          setError('迁移已取消，已移动的文件已恢复到原位置');
          setState('idle');
          // Refresh state so the UI reflects the restored folders.
          loadData({ silent: true });
        } else {
          setError(
            `迁移失败：${event.error || '未知错误'}${
              event.action ? `（${event.action}）` : ''
            }`,
          );
          setState('failed');
          // Refresh folder state so folders that completed before the failure
          // expose rollback actions.
          GetFolders().then(setFolders).catch(() => {});
        }
      }
    });

    return () => {
      unsubProgress();
      unsubDone();
    };
  }, [state, selectedDisk, loadData]);

  const handleRollback = async (folderName: string) => {
    try {
      await Rollback(folderName);
      setFolders(prev =>
        prev.map(f =>
          f.name === folderName ? { ...f, status: 'default', targetPath: '' } : f,
        ),
      );
    } catch (err) {
      setError(`回滚失败：${err instanceof Error ? err.message : '未知错误'}`);
    }
  };

  const handleDone = () => {
    setState('idle');
    loadData();
  };

  const totalSize = folders.reduce((sum, f) => sum + f.sizeBytes, 0);

  return (
    <div className="min-h-screen bg-surface font-sans text-text-primary">
      {/* Header */}
      <header className="bg-card border-b border-border px-6 py-4 shadow-fluent flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">MoveKit</h1>
          <p className="text-sm text-text-secondary">Windows 用户文件夹迁移工具</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowDefenderInfo(true)}
            className="w-8 h-8 rounded-fluent flex items-center justify-center text-text-secondary hover:bg-surface-alt transition-colors"
            title="安全提示"
            aria-label="安全提示"
          >
            🛡️
          </button>
          <button
            onClick={() => setShowHelpPanel(true)}
            className="w-8 h-8 rounded-fluent flex items-center justify-center text-text-secondary hover:bg-surface-alt transition-colors"
            title="帮助"
            aria-label="帮助"
          >
            ❓
          </button>
        </div>
      </header>

      {/* Main content */}
      <main className="max-w-2xl mx-auto p-6 space-y-8">
        {/* Error banner */}
        {error && (
          <div className="p-4 rounded-fluent bg-danger/10 border border-danger/20 flex items-start gap-3">
            <span className="text-danger mt-0.5">✕</span>
            <div className="flex-1">
              <p className="text-sm text-danger">{error}</p>
            </div>
            <button
              onClick={() => { setError(''); setState('idle'); loadData(); }}
              className="text-xs text-danger hover:underline flex-shrink-0"
            >
              关闭
            </button>
          </div>
        )}

        {/* State: checking */}
        {state === 'checking' && (
          <div className="text-center py-12">
            <div className="animate-spin w-8 h-8 border-2 border-accent border-t-transparent rounded-full mx-auto" />
            <p className="mt-4 text-text-secondary">正在检查系统状态…</p>
          </div>
        )}

        {/* State: idle — show folders + disk selector */}
        {(state === 'idle' || state === 'confirming') && (
          <>
            <FolderList folders={folders} />

            <SectionErrorBoundary>
            <div className="p-4 rounded-fluent bg-card border border-border">
              <DiskSelector disks={disks} selected={selectedDisk} onSelect={setSelectedDisk} />
            </div>
            </SectionErrorBoundary>

            {totalSize > 0 && selectedDisk && (
              <div className="p-4 rounded-fluent bg-accent/5 border border-accent/10">
                <p className="text-sm text-text-secondary">
                  将迁移 <span className="font-medium text-text-primary">{folders.length}</span> 个文件夹，
                  共计 <span className="font-medium text-text-primary">{(totalSize / 1e9).toFixed(1)} GB</span> 到{' '}
                  <span className="font-medium text-accent">{selectedDisk}</span>
                </p>
              </div>
            )}

            {/* Confirm dialog */}
            {state === 'confirming' ? (
              <div className="p-6 rounded-fluent bg-card border border-warning/30 shadow-fluent-lg space-y-4">
                <div className="flex items-center gap-2">
                  <span className="text-warning text-lg">⚠</span>
                  <h3 className="font-semibold text-text-primary">确认迁移</h3>
                </div>
                <p className="text-sm text-text-secondary">
                  即将把 {folders.length} 个文件夹从 C 盘迁移到 {selectedDisk}。
                  迁移过程中请勿关闭程序。确认继续？
                </p>
                <div className="flex gap-3">
                  <button
                    onClick={() => setState('idle')}
                    className="px-4 py-2 text-sm text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
                  >
                    取消
                  </button>
                  <button
                    onClick={handleConfirm}
                    className="px-4 py-2 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
                  >
                    确认迁移
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={handleStart}
                disabled={!selectedDisk || folders.length === 0}
                className={`w-full py-3 text-sm font-medium rounded-fluent transition-colors ${
                  selectedDisk && folders.length > 0
                    ? 'bg-accent hover:bg-accent-dark text-white'
                    : 'bg-surface-alt text-text-disabled cursor-not-allowed'
                }`}
              >
                开始迁移
              </button>
            )}
          </>
        )}

        {/* State: migrating */}
        {state === 'migrating' && (
          <ProgressBar progress={progress} onCancel={handleCancel} />
        )}

        {/* State: completed / failed */}
        {(state === 'completed' || state === 'failed') && (
          <ResultPage
            success={state === 'completed'}
            folders={folders}
            freedBytes={freedBytes}
            onRollback={handleRollback}
            onDone={handleDone}
          />
        )}
      </main>

      {/* ── Dialog overlays ── */}

      {showOneDriveWarning && (
        <OneDriveWarning
          folders={oneDriveFolders}
          onClose={() => setShowOneDriveWarning(false)}
        />
      )}

      {showDefenderInfo && (
        <DefenderInfo onClose={() => setShowDefenderInfo(false)} />
      )}

      {showFirstRunGuide && (
        <FirstRunGuide
          onComplete={() => {
            setShowFirstRunGuide(false);
            localStorage.setItem('movekit-first-run-seen', 'true');
          }}
        />
      )}

      <HelpPanel
        open={showHelpPanel}
        onClose={() => setShowHelpPanel(false)}
      />
    </div>
  );
}

// Root export: wrap with ErrorBoundary
export default function App() {
  return (
    <ErrorBoundary>
      <AppInner />
    </ErrorBoundary>
  );
}
