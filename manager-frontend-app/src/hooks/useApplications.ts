import { useEffect, useMemo, useState } from 'react'
import { appManagerApi } from '../api'
import type { Application } from '../types'

type UseApplicationsState = {
  loading: boolean
  error: string | null
  applications: Application[]
}

export function useApplications(): UseApplicationsState {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [applications, setApplications] = useState<Application[]>([])

  useEffect(() => {
    let alive = true

    async function load() {
      try {
        setLoading(true)
        const response = await appManagerApi.listApplications()

        if (!alive) {
          return
        }

        setApplications(response.data.items ?? [])
        setError(null)
      } catch (err) {
        if (!alive) {
          return
        }

        setError(err instanceof Error ? err.message : '加载应用列表失败')
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

  return useMemo(() => ({ loading, error, applications }), [loading, error, applications])
}
