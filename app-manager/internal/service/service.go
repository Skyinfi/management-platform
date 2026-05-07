package service

import (
	"strconv"

	"github.com/Skyinfi/app-manager/internal/model"
	"github.com/Skyinfi/app-manager/internal/store"
)

type Service struct {
	store *store.Store
}

func New(st *store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) Dashboard() model.DashboardResponse {
	apps := s.store.ListApplications()
	return model.DashboardResponse{
		Metrics:      buildMetrics(apps),
		Activities:   s.store.ListActivities(),
		Applications: apps,
	}
}

func (s *Service) Applications() []model.Application {
	return s.store.ListApplications()
}

func (s *Service) Logs(name string) []string {
	return s.store.ListLogs(name)
}

func (s *Service) Action(name, action string) (string, bool) {
	return s.store.ApplyAction(name, action)
}

func buildMetrics(apps []model.Application) []model.MetricItem {
	running, stopped, warning := 0, 0, 0
	for _, app := range apps {
		switch app.Status {
		case "running":
			running++
		case "stopped":
			stopped++
		case "warning":
			warning++
		}
	}
	return []model.MetricItem{
		{Label: "运行中", Value: strconv.Itoa(running), Delta: "今日新增 4 个"},
		{Label: "已停止", Value: strconv.Itoa(stopped), Delta: "其中 2 个待处理"},
		{Label: "告警中", Value: strconv.Itoa(warning), Delta: "需要关注"},
		{Label: "平均延迟", Value: "184ms", Delta: "较昨日降低 12ms"},
	}
}
