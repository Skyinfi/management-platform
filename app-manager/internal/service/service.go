package service

import (
	"context"
	"strconv"

	"github.com/Skyinfi/management-platform/app-manager/internal/model"
	"github.com/Skyinfi/management-platform/app-manager/internal/store"
)

type Service struct {
	store   *store.Store
	docker  *DockerService
	process *ProcessService
}

func New(st *store.Store, opts ...ServiceOption) *Service {
	svc := &Service{store: st}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

type ServiceOption func(*Service)

func WithDocker(d *DockerService) ServiceOption {
	return func(s *Service) { s.docker = d }
}

func WithProcess(p *ProcessService) ServiceOption {
	return func(s *Service) { s.process = p }
}

func (s *Service) Dashboard() model.DashboardResponse {
	apps := s.Applications()
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

func (s *Service) AllApplications(ctx context.Context) []model.Application {
	var apps []model.Application

	if s.docker != nil && s.docker.Enabled() {
		dockerApps := s.docker.Applications(ctx)
		apps = append(apps, dockerApps...)
	}

	if s.process != nil {
		processApps := s.process.Applications(ctx)
		apps = append(apps, processApps...)
	}

	if len(apps) == 0 {
		apps = s.store.ListApplications()
	}

	return apps
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
