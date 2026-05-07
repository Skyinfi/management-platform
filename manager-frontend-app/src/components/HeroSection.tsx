import type { MetricItem } from '../types'

type Props = {
  metrics: MetricItem[]
}

export function HeroSection({ metrics }: Props) {
  return (
    <section className="hero" id="dashboard">
      <div>
        <span className="eyebrow">应用管理平台</span>
        <h1>在一个地方统一监控、操作和查看所有已部署应用。</h1>
        <p className="hero-copy">
          这是一个基于设计书实现的前端原型，使用 mock 数据先验证页面结构、导航和交互形式，
          方便后续再无缝接入后端接口。
        </p>

        <div className="hero-actions">
          <button type="button" className="primary-btn">
            新增应用
          </button>
          <button type="button" className="secondary-btn">
            查看审计
          </button>
        </div>
      </div>

      <div className="hero-panel">
        <div className="panel-title">实时总览</div>
        <div className="metric-grid">
          {metrics.map((item) => (
            <article key={item.label} className="metric-card">
              <span>{item.label}</span>
              <strong>{item.value}</strong>
              <small>{item.delta}</small>
            </article>
          ))}
        </div>
      </div>
    </section>
  )
}
