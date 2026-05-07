import type { ActivityItem } from '../types'

type Props = {
  title: string
  activities: ActivityItem[]
  onSelectLogApp?: (appName: string) => void
}

export function ActivityTimeline({ title, activities, onSelectLogApp }: Props) {
  return (
    <section className="section" aria-label={title}>
      <div className="section-header compact">
        <div>
          <span className="section-label">操作日志</span>
          <h2>{title}</h2>
        </div>
      </div>

      <div className="timeline">
        {activities.map((item) => (
          <div key={item.time + item.title} className={`timeline-item tone-${item.tone}`}>
            <span className="timeline-time">{item.time}</span>
            <div>
              <strong>{item.title}</strong>
              <p>{item.detail}</p>
              <button
                type="button"
                className="text-link"
                onClick={() => onSelectLogApp?.(item.appName)}
              >
                查看相关日志
              </button>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
