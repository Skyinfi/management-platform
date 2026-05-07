import { useEffect, useMemo, useState } from 'react'
import { appManagerApi } from '../api'

export type LogEntry = {
  id: string
  line: string
}

type UseAppLogsState = {
  loading: boolean
  error: string | null
  selectedApp: string
  logs: LogEntry[]
  loadLogs: (appName: string) => Promise<void>
}

const fallbackLogs = [
  '2026-05-07 09:43:12 [INFO] container started successfully',
  '2026-05-07 09:43:13 [INFO] health check passed',
  '2026-05-07 09:43:14 [INFO] listening on port 8080',
]

const fallbackApp = 'manager-api'

export function useAppLogs(initialAppName = fallbackApp): UseAppLogsState {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedApp, setSelectedApp] = useState(initialAppName)
  const [logs, setLogs] = useState<LogEntry[]>(
    fallbackLogs.map((line, index) => ({ id: `${index}-${line}`, line })),
  )

  async function loadLogs(appName: string) {
    setSelectedApp(appName)
    setLoading(true)

    try {
      const response = await appManagerApi.fetchApplicationLogs(appName)
      setLogs(response.data.lines.map((line: string, index: number) => ({ id: `${appName}-${index}`, line })))
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载日志失败')
      setLogs(fallbackLogs.map((line, index) => ({ id: `${appName}-${index}`, line })))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadLogs(initialAppName)
  }, [initialAppName])

  return useMemo(
    () => ({ loading, error, selectedApp, logs, loadLogs }),
    [loading, error, selectedApp, logs],
  )
}
