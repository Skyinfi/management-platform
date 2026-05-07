import type { Application } from '../types'
import { StatusPill } from './StatusPill'

type Props = {
  title: string
  type: Application['type']
  applications: Application[]
  pendingAction?: string | null
  onStart?: (name: string) => void
  onStop?: (name: string) => void
  onRestart?: (name: string) => void
}

export function ApplicationTable({ title, type, applications, pendingAction, onStart, onStop, onRestart }: Props) {
  return (
    <section className="section" aria-label={title}>
      <div className="section-header">
        <div>
          <span className="section-label">{type === 'Docker' ? 'Docker 应用' : '进程应用'}</span>
          <h2>{title}</h2>
        </div>
        <button type="button" className="ghost-btn">
          刷新
        </button>
      </div>

      <div className="table-card">
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>类型</th>
              <th>状态</th>
              <th>版本</th>
              <th>地址</th>
              <th>负责人</th>
              <th>CPU</th>
              <th>内存</th>
              <th>运行时长</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {applications.map((item) => (
              <tr key={item.name}>
                <td className="name-cell">{item.name}</td>
                <td>{item.type === 'Docker' ? 'Docker' : '进程'}</td>
                <td>
                  <StatusPill status={item.status} />
                </td>
                <td>{item.version}</td>
                <td>{item.endpoint}</td>
                <td>{item.owner}</td>
                <td>{item.cpu}</td>
                <td>{item.memory}</td>
                <td>{item.uptime}</td>
                <td>{item.updatedAt}</td>
                <td>
                  <div className="actions">
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
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
