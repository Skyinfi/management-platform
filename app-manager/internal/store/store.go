package store

import (
	"sort"
	"sync"

	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

type Store struct {
	mu           sync.RWMutex
	applications []model.Application
	activities   []model.ActivityItem
	logs         map[string][]string
}

func Default() *Store {
	return &Store{
		applications: []model.Application{
			{Name: "manager-api", Type: "Docker", Status: "running", Version: "v2.8.1", Endpoint: "10.0.1.12:8080", Owner: "平台团队", CPU: "12.4%", Memory: "384 MB", Uptime: "12d 04h", UpdatedAt: "2 分钟前"},
			{Name: "order-service", Type: "Process", Status: "running", Version: "v1.4.0", Endpoint: "10.0.1.12:9001", Owner: "交易团队", CPU: "8.2%", Memory: "228 MB", Uptime: "8d 11h", UpdatedAt: "8 分钟前"},
			{Name: "nginx-gateway", Type: "Docker", Status: "warning", Version: "v1.25.3", Endpoint: "10.0.1.12:80", Owner: "基础设施", CPU: "3.1%", Memory: "92 MB", Uptime: "30d 02h", UpdatedAt: "1 分钟前"},
			{Name: "report-worker", Type: "Process", Status: "stopped", Version: "v0.9.7", Endpoint: "batch-only", Owner: "数据团队", CPU: "0%", Memory: "0 MB", Uptime: "—", UpdatedAt: "23 分钟前"},
		},
		activities: []model.ActivityItem{
			{Time: "09:42", Title: "manager-api 已重启", Detail: "运维人员 Mike 执行了滚动重启", Tone: "success", AppName: "manager-api"},
			{Time: "09:31", Title: "nginx-gateway 出现健康告警", Detail: "容器 CPU 峰值升高", Tone: "warning", AppName: "nginx-gateway"},
			{Time: "09:18", Title: "order-service 已更新配置", Detail: "新的环境变量已发布", Tone: "info", AppName: "order-service"},
			{Time: "08:56", Title: "report-worker 已手动停止", Detail: "批处理任务暂停维护", Tone: "muted", AppName: "report-worker"},
		},
		logs: map[string][]string{
			"manager-api":   {"2026-05-07 09:43:12 [INFO] container started successfully", "2026-05-07 09:43:13 [INFO] health check passed", "2026-05-07 09:43:14 [INFO] listening on port 8080"},
			"order-service": {"2026-05-07 09:41:12 [INFO] service start request received", "2026-05-07 09:41:13 [INFO] database connection ok"},
			"nginx-gateway": {"2026-05-07 09:31:12 [WARN] upstream latency increased", "2026-05-07 09:31:14 [INFO] retrying health probe"},
			"report-worker": {"2026-05-07 08:56:12 [INFO] job paused by operator", "2026-05-07 08:56:14 [INFO] worker gracefully stopped"},
		},
	}
}

func (s *Store) ListApplications() []model.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]model.Application(nil), s.applications...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Store) ListActivities() []model.ActivityItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]model.ActivityItem(nil), s.activities...)
}

func (s *Store) ListLogs(name string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.logs[name]...)
}

func (s *Store) ApplyAction(name, action string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.applications {
		if s.applications[i].Name != name {
			continue
		}
		switch action {
		case "start":
			s.applications[i].Status = "running"
			s.applications[i].UpdatedAt = "刚刚"
			return "已启动", true
		case "stop":
			s.applications[i].Status = "stopped"
			s.applications[i].UpdatedAt = "刚刚"
			return "已停止", true
		case "restart":
			s.applications[i].Status = "running"
			s.applications[i].UpdatedAt = "刚刚"
			return "已重启", true
		default:
			return "不支持的操作", true
		}
	}
	return "", false
}
