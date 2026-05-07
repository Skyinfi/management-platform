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
		applications: []model.Application{},
		activities:   []model.ActivityItem{},
		logs:         map[string][]string{},
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
