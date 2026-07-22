import { useState, useEffect, useCallback, useRef } from 'react';
import { AppState, DiskInfo, FolderInfo, ProgressInfo } from './types';
import { GetFolders, GetDisks, StartMigration, Rollback, CheckOneDriveKFM, GetVersion, CancelMigration, DetectConflicts, onMigrationProgress, onMigrationDone } from './lib/wails';
import DiskSelector from './components/DiskSelector';
import FolderList from './components/FolderList';
import ProgressBar from './components/ProgressBar';
import ResultPage from './components/ResultPage';
import OneDriveWarning from './components/OneDriveWarning';
import ConflictDialog, { type ConflictFile } from './components/ConflictDialog';
import CloudWarning from './components/CloudWarning';
import DefenderInfo from './components/DefenderInfo';
import UpdateNotice from './components/UpdateNotice';
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
  const [showConflictDialog, setShowConflictDialog] = useState(false);
  const [conflicts, setConflicts] = useState<ConflictFile[]>([]);
  const [showCloudWarning, setShowCloudWarning] = useState(false);
  const [cloudProviders] = useState<string[]>([]);
  const [showDefenderInfo, setShowDefenderInfo] = useState(false);
  const [showUpdateNotice, setShowUpdateNotice] = useState(false);
  const [updateInfo, setUpdateInfo] = useState({ currentVersion: '', newVersion: '', downloadUrl: '' });
  const [showFirstRunGuide, setShowFirstRunGuide] = useState(false);
  const [showHelpPanel, setShowHelpPanel] = useState(false);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);

  // Check first-run on mount
  useEffect(() => {
    const hasSeenGuide = localStorage.getItem('movekit-first-run-seen');
    if (!hasSeenGuide) {
      setShowFirstRunGuide(true);
    }
  }, []);

  // Check for updates on mount
  useEffect(() => {
    (async () => {
      try {
        const ver = await GetVersion();
        // In production, compare with remote latest version
        // For now, just store current version
        setUpdateInfo(prev => ({ ...prev, currentVersion: ver }));
      } catch {
        // ignore
      }
    })();
  }, []);

  // Heartbeat during migration — uses the useHeartbeat hook from ErrorBoundary
  const handleDead = useCallback(() => {
    setError('后端进程失去响应');
    setState('failed');
  }, []);
  useHeartbeat({ enabled: state === 'migrating', onDead: handleDead });

  const loadData = useCallback(async () => {
    setState('checking');
    try {
      const [f, d] = await Promise.all([GetFolders(), GetDisks()]);
      setFolders(f);
      setDisks(d);
      setState('idle');
    } catch {
      setError('无法加载数据，请检查后端服务');
      setState('failed');
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
    // Check for file conflicts before migrating.
    try {
      const allConflicts: ConflictFile[] = [];
      for (const f of folders) {
        if (f.status === 'migrated' || f.status === 'not_on_c') continue;
        const target = `${selectedDisk}\\${f.name}`;
        const result = await DetectConflicts(f.sourcePath, target);
        if (result && result.length > 0) {
          for (const c of result) {
            allConflicts.push({
              filename: c.relativePath || c.RelativePath || '',
              sourceSize: c.srcSize || c.SrcSize || 0,
              sourceDate: c.srcModTime || c.SrcModTime || '',
              targetSize: c.dstSize || c.DstSize || 0,
              targetDate: c.dstModTime || c.DstModTime || '',
            });
          }
        }
      }
      if (allConflicts.length > 0) {
        setConflicts(allConflicts);
        setShowConflictDialog(true);
        return;
      }
    } catch {
      // If conflict detection fails, proceed anyway
    }

    setState('migrating');
    try {
      await StartMigration(selectedDisk);
    } catch (err) {
      const msg = err instanceof Error ? err.message : '未知错误';
      setError(msg.includes('syscall') || msg.includes('access denied')
        ? '迁移失败：权限不足，请以管理员身份运行 MoveKit 后重试。'
        : `迁移失败：${msg}`);
      setState('failed');
    }
  };

  const handleCancel = () => {
    setShowCancelConfirm(true);
  };

  const handleCancelConfirm = async () => {
    setShowCancelConfirm(false);
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

  // ── Wails Event Subscriptions (real progress + completion) ──
  useEffect(() => {
    if (state !== 'migrating') return;

    const unsubProgress = onMigrationProgress((info) => {
      setProgress(info);
    });

    const unsubDone = onMigrationDone((event) => {
      if (event.success) {
        setFolders(prev => prev.map(f => ({ ...f, status: 'migrated' as const, targetPath: f.sourcePath.replace('C:', selectedDisk) })));
        setFreedBytes(progressRef.current.totalBytes);
        setState('completed');
      } else {
        setError(`迁移失败：${event.error || '未知错误'}`);
        setState('failed');
      }
    });

    return () => {
      unsubProgress();
      unsubDone();
    };
  }, [state, selectedDisk]);

  const handleRollback = async (folderName: string) => {
    try {
      await Rollback(folderName);
      setFolders(prev =>
        prev.map(f =>
          f.name === folderName ? { ...f, status: 'default', targetPath: '' } : f,
        ),
      );
    } catch (err) {
      const msg = err instanceof Error ? err.message : '未知错误';
      setError(msg.includes('syscall') || msg.includes('access denied')
        ? '回滚失败：权限不足，请以管理员身份运行 MoveKit 后重试。'
        : `回滚失败：${msg}`);
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

            <div className="p-4 rounded-fluent bg-card border border-border">
              <DiskSelector disks={disks} selected={selectedDisk} onSelect={setSelectedDisk} />
            </div>

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

      {showConflictDialog && (
        <ConflictDialog
          conflicts={conflicts}
          onSkip={() => {
            setShowConflictDialog(false);
            // Proceed with migration — conflicts will be skipped.
            setState('migrating');
            StartMigration(selectedDisk, 'skip').catch((err) => {
              const msg = err instanceof Error ? err.message : '未知错误';
              setError(`迁移失败：${msg}`);
              setState('failed');
            });
          }}
          onOverwrite={() => {
            setShowConflictDialog(false);
            // Proceed with migration — conflicts will be overwritten.
            setState('migrating');
            StartMigration(selectedDisk, 'overwrite').catch((err) => {
              const msg = err instanceof Error ? err.message : '未知错误';
              setError(`迁移失败：${msg}`);
              setState('failed');
            });
          }}
          onCancel={() => {
            setShowConflictDialog(false);
            setState('idle');
          }}
        />
      )}

      {showCloudWarning && (
        <CloudWarning
          providers={cloudProviders}
          onClose={() => setShowCloudWarning(false)}
        />
      )}

      {showDefenderInfo && (
        <DefenderInfo onClose={() => setShowDefenderInfo(false)} />
      )}

      {showUpdateNotice && (
        <UpdateNotice
          currentVersion={updateInfo.currentVersion}
          newVersion={updateInfo.newVersion}
          downloadUrl={updateInfo.downloadUrl}
          onLater={() => setShowUpdateNotice(false)}
        />
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

      {showCancelConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-card rounded-fluent shadow-fluent-lg border border-border max-w-md w-full mx-4 overflow-hidden">
            <div className="px-6 pt-6 pb-2">
              <div className="flex items-center gap-3">
                <span className="text-2xl">⚠️</span>
                <h2 className="text-lg font-semibold text-text-primary">确认取消迁移？</h2>
              </div>
              <p className="text-sm text-text-secondary mt-2">
                取消迁移可能导致部分文件已复制但未完成。已复制的文件不会自动清理。
              </p>
            </div>
            <div className="px-6 py-4 border-t border-border flex justify-end gap-3">
              <button
                onClick={() => setShowCancelConfirm(false)}
                className="px-4 py-2 text-sm text-text-secondary rounded-fluent border border-border hover:bg-surface-alt transition-colors"
              >
                继续迁移
              </button>
              <button
                onClick={handleCancelConfirm}
                className="px-4 py-2 text-sm font-medium text-white rounded-fluent bg-danger hover:bg-danger/90 transition-colors"
              >
                确认取消
              </button>
            </div>
          </div>
        </div>
      )}
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
