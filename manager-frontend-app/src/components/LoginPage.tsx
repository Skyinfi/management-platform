import { useEffect, useState } from 'react'

type Props = {
  loading?: boolean
  error?: string | null
  onLogin: (username: string, password: string, remember: boolean) => Promise<void>
}

const LAST_USERNAME_KEY = 'app-manager-last-username'
const REMEMBER_ME_KEY = 'app-manager-remember-me'

export function LoginPage({ loading = false, error = null, onLogin }: Props) {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin123')
  const [remember, setRemember] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [shake, setShake] = useState(false)

  useEffect(() => {
    const savedUsername = localStorage.getItem(LAST_USERNAME_KEY)
    const savedRemember = localStorage.getItem(REMEMBER_ME_KEY)

    if (savedRemember !== null) {
      setRemember(savedRemember === 'true')
    }

    if (savedUsername && savedRemember === 'true') {
      setUsername(savedUsername)
    }
  }, [])

  useEffect(() => {
    if (!error) {
      return
    }

    setShake(true)
    const timer = window.setTimeout(() => setShake(false), 500)
    return () => window.clearTimeout(timer)
  }, [error])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)

    try {
      if (remember) {
        localStorage.setItem(LAST_USERNAME_KEY, username)
        localStorage.setItem(REMEMBER_ME_KEY, 'true')
      } else {
        localStorage.removeItem(LAST_USERNAME_KEY)
        localStorage.setItem(REMEMBER_ME_KEY, 'false')
      }

      await onLogin(username, password, remember)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="login-shell">
      <div className={`login-card ${shake ? 'shake' : ''}`}>
        <span className="eyebrow">App Manager</span>
        <h1>登录到应用管理平台</h1>
        <p>先通过账号进入系统，再查看容器、进程、日志和运维状态。</p>

        <form className="login-form" onSubmit={handleSubmit}>
          <label>
            <span>用户名</span>
            <input value={username} onChange={(e) => setUsername(e.target.value)} />
          </label>
          <label>
            <span>密码</span>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
          </label>
          <label className="remember-row">
            <input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)} />
            <span>记住我</span>
          </label>
          {error ? <div className="state-banner state-banner-error">{error}</div> : null}
          <button type="submit" className="primary-btn" disabled={loading || submitting}>
            {loading || submitting ? '登录中...' : '登录'}
          </button>
        </form>
      </div>
    </div>
  )
}
