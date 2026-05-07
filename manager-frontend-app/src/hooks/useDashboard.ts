import { useEffect, useMemo, useState } from 'react'
import { appManagerApi } from '../api'
import type { ActivityItem, MetricItem } from '../types'
import { activity as mockActivity, metrics as mockMetrics } from '../data/mockData'

type UseDashboardState = {
  loading: boolean
  error: string | null
  metrics: MetricItem[]
  activity: ActivityItem[]
}

export function useDashboard(): UseDashboardState {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [metrics, setMetrics] = useState<MetricItem[]>(mockMetrics)
  const [activity, setActivity] = useState<ActivityItem[]>(mockActivity)

  useEffect(() => {
    let alive = true

    async function load() {
      try {
        setLoading(true)
        const response = await appManagerApi.getDashboard()

        if (!alive) {
          return
        }

        setMetrics(response.data.metrics)
        setActivity(response.data.activities)
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

  return useMemo(() => ({ loading, error, metrics, activity }), [loading, error, metrics, activity])
}
