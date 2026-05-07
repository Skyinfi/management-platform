import { useMemo, useState } from 'react'
import type { LogEntry } from '../hooks/useAppLogs'
import type { Application } from '../types'

type Props = {
  applications: Application[]
  selectedApp: string
  loading: boolean
  error: string | null
  logs: LogEntry[]
  onSelectApp: (appName: string) => void
  onRefresh: () => void
}

export function LogSidebar({ applications, selectedApp, loading, error, logs, onSelectApp, onRefresh }: Props) {
  const [query, setQuery] = useState('')

  const filteredApplications = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase()

    if (!normalizedQuery) {
      return applications
    }

    return applications.filter(
      (item) =>
        item.name.toLowerCase().includes(normalizedQuery) ||
        item.owner.toLowerCase().includes(normalizedQuery) ||
        item.type.toLowerCase().includes(normalizedQuery),
    )
  }, [applications, query])

  const visibleApplications = filteredApplications.some((item) => item.name === selectedApp)
    ? filteredApplications
    : applications.filter((item) => item.name === selectedApp).concat(filteredApplications)

  return (
    <aside className="log-sidebar" aria-label="应用日志侧边面板">
      <div className="log-sidebar-header">
        <div>
          <span className="section-label">日志面板</span>
          <h2>切换应用日志</h2>
        </div>
        <button type="button" className="ghost-btn" onClick={onRefresh}>
          刷新
        </button>
      </div>

      <label className="log-search">
        <span className="log-search-label">搜索应用</span>
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="输入应用名、负责人或类型"
        />
      </label>

      <div className="log-switcher">
        {visibleApplications.map((item) => (
          <button
            key={item.name}
            type="button"
            className={`log-switcher-item ${selectedApp === item.name ? 'active' : ''}`}
            onClick={() => onSelectApp(item.name)}
          >
            <strong>{item.name}</strong>
            <span>{item.type === 'Docker' ? 'Docker' : '进程'} · {item.owner}</span>
          </button>
        ))}
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
    </aside>
  )
}
