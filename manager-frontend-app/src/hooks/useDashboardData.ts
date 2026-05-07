import { useEffect, useMemo, useState } from 'react'
import { appManagerApi } from '../api'
import type { ActivityItem, Application, MetricItem } from '../types'

type DashboardState = {
  loading: boolean
  error: string | null
  metrics: MetricItem[]
  applications: Application[]
  activity: ActivityItem[]
}

export function useDashboardData(): DashboardState {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [metrics, setMetrics] = useState<MetricItem[]>([])
  const [applications, setApplications] = useState<Application[]>([])
  const [activity, setActivity] = useState<ActivityItem[]>([])

  useEffect(() => {
    let alive = true

    async function load() {
      try {
        setLoading(true)
        const [dashboardRes, applicationsRes] = await Promise.all([
          appManagerApi.getDashboard(),
          appManagerApi.listApplications(),
        ])

        if (!alive) {
          return
        }

        setMetrics(dashboardRes.data.metrics)
        setActivity(dashboardRes.data.activities)
        setApplications(applicationsRes.data.items)
        setError(null)
      } catch (err) {
        if (!alive) {
          return
        }

        setError(err instanceof Error ? err.message : '加载仪表盘失败')
      } finally {
        if (alive) {
          setLoading(false)
        }
      }
    }

    void load()

    return () => {
      alive = false
    }
  }, [])

  return useMemo(
    () => ({ loading, error, metrics, applications, activity }),
    [loading, error, metrics, applications, activity],
  )
}
