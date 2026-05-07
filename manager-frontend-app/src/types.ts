export type AppStatus = 'running' | 'stopped' | 'warning'

export type ApplicationType = 'Docker' | 'Process'

export type Application = {
  name: string
  type: ApplicationType
  status: AppStatus
  version: string
  endpoint: string
  owner: string
  cpu: string
  memory: string
  uptime: string
  updatedAt: string
}

export type DiscoveredProcess = {
  id: string
  name: string
  pid: number
  user: string
  endpoint: string
  protocol: string
  command: string
  exePath: string
  cwd: string
  managed: boolean
  inDocker: boolean
  adoptable: boolean
}

export type ActivityTone = 'success' | 'warning' | 'info' | 'muted'

export type ActivityItem = {
  time: string
  title: string
  detail: string
  tone: ActivityTone
  appName: string
}

export type MetricItem = {
  label: string
  value: string
  delta: string
}
