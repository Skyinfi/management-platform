type Props = {
  userName?: string | null
  onLogout: () => void
}

export function Sidebar({ userName, onLogout }: Props) {
  return (
    <aside className="sidebar">
      <div>
        <div className="brand">App Manager</div>
        <p className="brand-copy">统一管理容器应用和进程应用的运维控制台。</p>
      </div>

      <nav className="menu">
        <a className="menu-item active" href="#dashboard">
          仪表盘
        </a>
        <a className="menu-item" href="#docker">
          Docker 应用
        </a>
        <a className="menu-item" href="#process">
          进程应用
        </a>
        <a className="menu-item" href="#activity">
          操作日志
        </a>
      </nav>

      <div className="sidebar-card">
        <div className="sidebar-card-title">当前用户</div>
        <div className="cluster-row">
          <span>登录账号</span>
          <strong>{userName ?? '管理员'}</strong>
        </div>
        <button type="button" className="logout-btn" onClick={onLogout}>
          退出登录
        </button>
      </div>

      <div className="sidebar-card">
        <div className="sidebar-card-title">集群状态</div>
        <div className="cluster-row">
          <span>主机</span>
          <strong>Ubuntu-01</strong>
        </div>
        <div className="cluster-row">
          <span>运行时长</span>
          <strong>43d 05h</strong>
        </div>
        <div className="cluster-row">
          <span>健康状态</span>
          <strong>稳定</strong>
        </div>
      </div>
    </aside>
  )
}
