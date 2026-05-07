import './App.css'
import { ActivityTimeline } from './components/ActivityTimeline'
import { ApplicationCards } from './components/ApplicationCards'
import { ApplicationTable } from './components/ApplicationTable'
import { HeroSection } from './components/HeroSection'
import { LogSidebar } from './components/LogSidebar'
import { LoginPage } from './components/LoginPage'
import { Sidebar } from './components/Sidebar'
import { useAppActions, useAppLogs, useApplications, useDashboard } from './hooks'
import { useAuth } from './auth/useAuth'

function App() {
  const auth = useAuth()
  const { metrics, activity, loading: dashboardLoading, error: dashboardError } = useDashboard()
  const { applications, loading: applicationsLoading, error: applicationsError } = useApplications()
  const { pendingAction, actionError, runAction } = useAppActions()
  const { loading: logsLoading, error: logsError, selectedApp, logs, loadLogs } = useAppLogs()
  const dockerApps = applications.filter((item) => item.type === 'Docker')
  const processApps = applications.filter((item) => item.type === 'Process')
  const error = dashboardError ?? applicationsError ?? actionError
  const loading = dashboardLoading || applicationsLoading

  if (!auth.ready) {
    return <div className="login-shell"><div className="login-card">正在检查登录状态...</div></div>
  }

  if (!auth.authenticated) {
    return <LoginPage error={auth.error} onLogin={auth.login} />
  }

  return (
    <div className="shell">
      <Sidebar userName={auth.user?.name} onLogout={auth.logout} />

      <main className="content">
        {loading ? <div className="state-banner">正在加载仪表盘数据...</div> : null}
        {error ? <div className="state-banner state-banner-error">{error}</div> : null}

        <HeroSection metrics={metrics} />
        <ApplicationTable
          title="容器舰队"
          type="Docker"
          applications={dockerApps}
          pendingAction={pendingAction}
          onStart={(name) => void runAction(name, 'start')}
          onStop={(name) => void runAction(name, 'stop')}
          onRestart={(name) => void runAction(name, 'restart')}
        />
        <section className="split" id="process">
          <div className="split-card">
            <ApplicationCards
              title="系统服务"
              applications={processApps}
              pendingAction={pendingAction}
              onStart={(name) => void runAction(name, 'start')}
              onStop={(name) => void runAction(name, 'stop')}
              onRestart={(name) => void runAction(name, 'restart')}
            />
          </div>
          <div className="split-card" id="activity">
            <ActivityTimeline title="运维时间线" activities={activity} onSelectLogApp={(name) => void loadLogs(name)} />
          </div>
          <LogSidebar
            applications={applications}
            selectedApp={selectedApp}
            loading={logsLoading}
            error={logsError}
            logs={logs}
            onSelectApp={(name) => void loadLogs(name)}
            onRefresh={() => void loadLogs(selectedApp)}
          />
        </section>
      </main>
    </div>
  )
}

export default App
