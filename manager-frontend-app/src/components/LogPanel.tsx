import type { LogEntry } from '../hooks/useAppLogs'

type Props = {
  appName: string
  loading: boolean
  error: string | null
  logs: LogEntry[]
  onRefresh: () => void
}

export function LogPanel({ appName, loading, error, logs, onRefresh }: Props) {
  return (
    <section className="section" aria-label="应用日志">
      <div className="section-header compact">
        <div>
          <span className="section-label">实时日志</span>
          <h2>{appName}</h2>
        </div>
        <button type="button" className="ghost-btn" onClick={onRefresh}>
          刷新日志
        </button>
      </div>

      <div className="log-panel">
        {loading ? <div className="state-banner">正在加载日志...</div> : null}
        {error ? <div className="state-banner state-banner-error">{error}</div> : null}

        {!loading && logs.length === 0 ? <div className="empty-state">暂无日志内容</div> : null}

        <pre className="log-output">
          {logs.map((entry) => (
            <div key={entry.id}>{entry.line}</div>
          ))}
        </pre>
      </div>
    </section>
  )
}
