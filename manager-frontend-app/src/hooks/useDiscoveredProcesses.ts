import { useMemo, useState } from 'react'
import { appManagerApi } from '../api'
import type { DiscoveredProcess } from '../types'

type UseDiscoveredProcessesState = {
  loading: boolean
  error: string | null
  scanned: boolean
  discoveredProcesses: DiscoveredProcess[]
  scanProcesses: () => Promise<void>
}

export function useDiscoveredProcesses(): UseDiscoveredProcessesState {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [scanned, setScanned] = useState(false)
  const [discoveredProcesses, setDiscoveredProcesses] = useState<DiscoveredProcess[]>([])

  async function scanProcesses() {
    try {
      setLoading(true)
      const response = await appManagerApi.discoverProcesses()
      setDiscoveredProcesses(response.data.items ?? [])
      setError(null)
      setScanned(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : '扫描监听端口失败')
      setScanned(true)
    } finally {
      setLoading(false)
    }
  }

  return useMemo(
    () => ({ loading, error, scanned, discoveredProcesses, scanProcesses }),
    [loading, error, scanned, discoveredProcesses],
  )
}
