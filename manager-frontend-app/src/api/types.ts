import type { ActivityItem, Application, MetricItem } from '../types'

export type ApiResponse<T> = {
  code: number
  message: string
  data: T
}

export type ListApplicationsResponse = {
  items: Application[]
}

export type DashboardResponse = {
  metrics: MetricItem[]
  activities: ActivityItem[]
  applications: Application[]
}

export type AppActionResponse = {
  success: boolean
  message: string
}

export type ApplicationLogResponse = {
  lines: string[]
}
