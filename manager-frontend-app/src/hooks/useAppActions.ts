import { useCallback, useState } from 'react'
import { appManagerApi } from '../api'

export function useAppActions() {
  const [pendingAction, setPendingAction] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const runAction = useCallback(async (name: string, action: 'start' | 'stop' | 'restart') => {
    setPendingAction(`${action}:${name}`)
    setActionError(null)

    try {
      if (action === 'start') {
        await appManagerApi.startApplication(name)
      }

      if (action === 'stop') {
        await appManagerApi.stopApplication(name)
      }

      if (action === 'restart') {
        await appManagerApi.restartApplication(name)
      }
    } catch (err) {
      setActionError(err instanceof Error ? err.message : '操作失败')
    } finally {
      setPendingAction(null)
    }
  }, [])

  return {
    pendingAction,
    actionError,
    runAction,
  }
}
