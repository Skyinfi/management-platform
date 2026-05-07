import type { AppStatus } from '../types'

type Props = {
  status: AppStatus
}

export function StatusPill({ status }: Props) {
  const text = status === 'running' ? '运行中' : status === 'stopped' ? '已停止' : '告警'

  return <span className={`pill pill-${status}`}>{text}</span>
}
