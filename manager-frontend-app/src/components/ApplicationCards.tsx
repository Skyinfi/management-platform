import type { Application } from '../types'
import { StatusPill } from './StatusPill'

type Props = {
  title: string
  applications: Application[]
  pendingAction?: string | null
  onStart?: (name: string) => void
  onStop?: (name: string) => void
  onRestart?: (name: string) => void
}

export function ApplicationCards({ title, applications, pendingAction, onStart, onStop, onRestart }: Props) {
  return (
    <section className="section" aria-label={title}>
      <div className="section-header compact">
        <div>
          <span className="section-label">进程应用</span>
          <h2>{title}</h2>
        </div>
      </div>

      <div className="stack-list">
        {applications.map((item) => (
          <article key={item.name} className="stack-item">
            <div>
              <div className="stack-head">
                <strong>{item.name}</strong>
                <StatusPill status={item.status} />
              </div>
              <p>{item.owner} · {item.version} · {item.endpoint}</p>
            </div>
            <div className="stack-meta">
              <span>CPU {item.cpu}</span>
              <span>内存 {item.memory}</span>
              <span>更新于 {item.updatedAt}</span>
            </div>
            <div className="actions stack-actions">
              <button
                type="button"
                disabled={pendingAction === `start:${item.name}`}
                onClick={() => onStart?.(item.name)}
              >
                启动
              </button>
              <button
                type="button"
                disabled={pendingAction === `stop:${item.name}`}
                onClick={() => onStop?.(item.name)}
              >
                停止
              </button>
              <button
                type="button"
                disabled={pendingAction === `restart:${item.name}`}
                onClick={() => onRestart?.(item.name)}
              >
                重启
              </button>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
