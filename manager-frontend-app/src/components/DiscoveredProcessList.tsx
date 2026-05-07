import type { DiscoveredProcess } from '../types'

type Props = {
  processes: DiscoveredProcess[]
}

export function DiscoveredProcessList({ processes }: Props) {
  const unmanaged = processes.filter((item) => !item.managed && !item.inDocker)

  if (processes.length === 0) {
    return null
  }

  return (
    <div className="discovery-list">
      <div className="discovery-summary">
        <span>发现 {unmanaged.length} 个未接管监听进程</span>
        <span>{processes.length} 个监听端口</span>
      </div>

      {unmanaged.length === 0 ? (
        <div className="empty-state">未发现新的宿主机监听进程</div>
      ) : (
        unmanaged.map((item) => (
          <article key={item.id} className="discovery-item">
            <div className="discovery-head">
              <strong>{item.name}</strong>
              <span className={item.adoptable ? 'adopt-pill' : 'adopt-pill muted'}>
                {item.adoptable ? '可生成 unit' : '需确认路径'}
              </span>
            </div>
            <div className="discovery-meta">
              <span>{item.endpoint}</span>
              <span>PID {item.pid}</span>
              <span>{item.user || '未知用户'}</span>
            </div>
            <p title={item.command}>{item.command || item.exePath || item.cwd}</p>
            {item.cwd ? <small>{item.cwd}</small> : null}
          </article>
        ))
      )}
    </div>
  )
}
