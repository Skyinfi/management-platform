import { useEffect, useMemo, useState } from 'react'
import { appManagerApi } from '../api'
import type { ActivityItem, Application, MetricItem } from '../types'
import { activity as mockActivity, applications as mockApplications, metrics as mockMetrics } from '../data/mockData'

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
  const [metrics, setMetrics] = useState<MetricItem[]>(mockMetrics)
  const [applications, setApplications] = useState<Application[]>(mockApplications)
  const [activity, setActivity] = useState<ActivityItem[]>(mockActivity)

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

        setMetrics(dashboardRes.data.data.metrics)
        setActivity(dashboardRes.data.data.activities)
        setApplications(applicationsRes.data.data.items)
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
