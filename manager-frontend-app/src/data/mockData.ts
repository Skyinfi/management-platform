import type { ActivityItem, Application, MetricItem } from '../types'

export const applications: Application[] = [
  {
    name: 'manager-api',
    type: 'Docker',
    status: 'running',
    version: 'v2.8.1',
    endpoint: '10.0.1.12:8080',
    owner: '平台团队',
    cpu: '12.4%',
    memory: '384 MB',
    uptime: '12d 04h',
    updatedAt: '2 分钟前',
  },
  {
    name: 'order-service',
    type: 'Process',
    status: 'running',
    version: 'v1.4.0',
    endpoint: '10.0.1.12:9001',
    owner: '交易团队',
    cpu: '8.2%',
    memory: '228 MB',
    uptime: '8d 11h',
    updatedAt: '8 分钟前',
  },
  {
    name: 'nginx-gateway',
    type: 'Docker',
    status: 'warning',
    version: 'v1.25.3',
    endpoint: '10.0.1.12:80',
    owner: '基础设施',
    cpu: '3.1%',
    memory: '92 MB',
    uptime: '30d 02h',
    updatedAt: '1 分钟前',
  },
  {
    name: 'report-worker',
    type: 'Process',
    status: 'stopped',
    version: 'v0.9.7',
    endpoint: 'batch-only',
    owner: '数据团队',
    cpu: '0%',
    memory: '0 MB',
    uptime: '—',
    updatedAt: '23 分钟前',
  },
]

export const metrics: MetricItem[] = [
  { label: '运行中', value: '28', delta: '今日新增 4 个' },
  { label: '已停止', value: '4', delta: '其中 2 个待处理' },
  { label: '告警中', value: '2', delta: '需要关注' },
  { label: '平均延迟', value: '184ms', delta: '较昨日降低 12ms' },
]

export const activity: ActivityItem[] = [
  {
    time: '09:42',
    title: 'manager-api 已重启',
    detail: '运维人员 Mike 执行了滚动重启',
    tone: 'success',
    appName: 'manager-api',
  },
  {
    time: '09:31',
    title: 'nginx-gateway 出现健康告警',
    detail: '容器 CPU 峰值升高',
    tone: 'warning',
    appName: 'nginx-gateway',
  },
  {
    time: '09:18',
    title: 'order-service 已更新配置',
    detail: '新的环境变量已发布',
    tone: 'info',
    appName: 'order-service',
  },
  {
    time: '08:56',
    title: 'report-worker 已手动停止',
    detail: '批处理任务暂停维护',
    tone: 'muted',
    appName: 'report-worker',
  },
]
