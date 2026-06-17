import { Component, type ReactNode } from 'react';
import { HealthCheck } from '../lib/wails';

// ── Error Boundary (class component for catch) ──

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('[MoveKit ErrorBoundary]', error, errorInfo);
  }

  private handleRestart = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-surface flex items-center justify-center p-6">
          <div className="max-w-md w-full bg-card rounded-fluent shadow-fluent-lg border border-border overflow-hidden">
            <div className="px-6 pt-6 pb-2">
              <div className="flex items-center gap-3">
                <span className="text-2xl">💥</span>
                <h2 className="text-lg font-semibold text-text-primary">
                  程序遇到了一个错误
                </h2>
              </div>
            </div>
            <div className="px-6 py-4 space-y-3">
              <p className="text-sm text-text-secondary">
                MoveKit 在渲染过程中发生了意外错误。这通常不会影响您的数据安全。
              </p>
              <div className="p-3 rounded-fluent bg-danger/5 border border-danger/20">
                <p className="text-xs text-danger font-mono break-all">
                  {this.state.error?.message ?? '未知错误'}
                </p>
              </div>
              <p className="text-xs text-text-secondary">
                您可以尝试重新启动程序来恢复。如果问题持续出现，请向我们反馈。
              </p>
            </div>
            <div className="px-6 py-4 border-t border-border flex justify-end">
              <button
                onClick={this.handleRestart}
                className="px-5 py-2 text-sm font-medium text-white rounded-fluent bg-accent hover:bg-accent-dark transition-colors"
              >
                重新启动
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// ── Heartbeat Hook ──

interface HeartbeatOptions {
  enabled: boolean;
  intervalMs?: number;
  maxFailures?: number;
  onDead: () => void;
}

import { useEffect, useRef } from 'react';

/**
 * Calls HealthCheck on an interval. After `maxFailures` consecutive failures
 * fires `onDead` (e.g. to show a disconnection banner).
 */
export function useHeartbeat({
  enabled,
  intervalMs = 5000,
  maxFailures = 3,
  onDead,
}: HeartbeatOptions) {
  const failuresRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setInterval>>(undefined);

  useEffect(() => {
    if (!enabled) {
      if (timerRef.current) clearInterval(timerRef.current);
      return;
    }

    failuresRef.current = 0;

    timerRef.current = setInterval(async () => {
      try {
        const ok = await HealthCheck();
        if (ok) {
          failuresRef.current = 0;
        } else {
          failuresRef.current += 1;
          if (failuresRef.current >= maxFailures) {
            onDead();
          }
        }
      } catch {
        failuresRef.current += 1;
        if (failuresRef.current >= maxFailures) {
          onDead();
        }
      }
    }, intervalMs);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [enabled, intervalMs, maxFailures, onDead]);
}

export default ErrorBoundary;
