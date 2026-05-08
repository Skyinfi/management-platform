package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Skyinfi/management-platform/app-manager/internal/middleware"
	"github.com/Skyinfi/management-platform/app-manager/internal/model"
	"github.com/Skyinfi/management-platform/app-manager/internal/service"
	"github.com/Skyinfi/management-platform/app-manager/internal/ws"
)

type Server struct {
	service  *service.Service
	auth     *service.AuthService
	docker   *service.DockerService
	process  *service.ProcessService
	audit    *service.AuditLog
	mux      *http.ServeMux
	security middleware.JWTValidator
	cors     bool
	origin   string
}

func New(
	svc *service.Service,
	auth *service.AuthService,
	docker *service.DockerService,
	process *service.ProcessService,
	audit *service.AuditLog,
	opts ...Option,
) *Server {
	s := &Server{
		service: svc,
		auth:    auth,
		docker:  docker,
		process: process,
		audit:   audit,
		mux:     http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.routes()
	return s
}

type Option func(*Server)

func WithJWTValidator(v middleware.JWTValidator) Option {
	return func(s *Server) { s.security = v }
}

func WithCORS(enabled bool, origin string) Option {
	return func(s *Server) {
		s.cors = enabled
		s.origin = origin
	}
}

func (s *Server) Routes() http.Handler {
	handler := http.Handler(s.mux)
	mws := []middleware.Middleware{middleware.Recovery(), middleware.Logging(), middleware.CORS(s.origin)}
	if s.security != nil {
		mws = append(mws, middleware.Auth(s.security))
	}
	return middleware.Chain(handler, mws...)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/auth/me", s.handleMe)
	s.mux.HandleFunc("/api/dashboard", s.handleDashboard)
	s.mux.HandleFunc("/api/applications", s.handleApplications)
	s.mux.HandleFunc("/api/applications/", s.handleApplicationAction)

	s.mux.HandleFunc("/api/docker/containers", s.handleDockerContainers)
	s.mux.HandleFunc("/api/docker/containers/", s.handleDockerContainerAction)
	s.mux.HandleFunc("/api/docker/images", s.handleDockerImages)

	s.mux.HandleFunc("/api/process/discovered", s.handleDiscoveredProcesses)
	s.mux.HandleFunc("/api/process/discovered/scan", s.handleTriggerScan)
	s.mux.HandleFunc("/api/process/discovered/", s.handleDiscoveredAction)
	s.mux.HandleFunc("/api/process/services", s.handleProcessServices)
	s.mux.HandleFunc("/api/process/services/", s.handleProcessServiceAction)

	s.mux.HandleFunc("/api/audit/logs", s.handleAuditLogs)

	s.mux.HandleFunc("/ws/docker/logs/", s.handleWSDockerLogs)
	s.mux.HandleFunc("/ws/process/logs/", s.handleWSProcessLogs)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    map[string]string{"status": "healthy"},
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    s.service.Dashboard(r.Context()),
	})
}

func (s *Server) handleApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.ListApplicationsResponse{Items: s.service.Applications(r.Context())},
	})
}

func (s *Server) handleApplicationAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/applications/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "not found"})
		return
	}

	name, action := parts[0], parts[1]
	if action == "logs" {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
			return
		}
		writeJSON(w, http.StatusOK, model.APIResponse{
			Code:    0,
			Message: "ok",
			Data:    model.ApplicationLogResponse{Lines: s.service.Logs(name)},
		})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	updated, ok := s.service.Action(name, action)
	if !ok {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "application not found"})
		return
	}

	operator := operatorFromContext(r.Context())
	s.audit.Record(operator, action, name, "application", updated)

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.AppActionResponse{Success: true, Message: updated},
	})
}

func (s *Server) handleDockerContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	if !s.docker.Enabled() {
		writeJSON(w, http.StatusOK, model.APIResponse{
			Code: 0, Message: "ok",
			Data: model.ContainerListResponse{Items: []model.Container{}},
		})
		return
	}

	containers, err := s.docker.ListContainers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.ContainerListResponse{Items: containers},
	})
}

func (s *Server) handleDockerContainerAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/docker/containers/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 1 {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "not found"})
		return
	}

	id := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodDelete {
			if err := s.docker.RemoveContainer(r.Context(), id); err != nil {
				writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
				return
			}
			operator := operatorFromContext(r.Context())
			s.audit.Record(operator, "delete", id, "docker", "已删除")
			writeJSON(w, http.StatusOK, model.APIResponse{
				Code: 0, Message: "ok",
				Data: model.AppActionResponse{Success: true, Message: "已删除"},
			})
			return
		}
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "not found"})
		return
	}

	action := parts[1]
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	var err error
	var msg string
	switch action {
	case "start":
		err = s.docker.StartContainer(r.Context(), id)
		msg = "已启动"
	case "stop":
		err = s.docker.StopContainer(r.Context(), id)
		msg = "已停止"
	case "restart":
		err = s.docker.RestartContainer(r.Context(), id)
		msg = "已重启"
	default:
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "unsupported action: " + action})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	operator := operatorFromContext(r.Context())
	s.audit.Record(operator, action, id, "docker", msg)

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code: 0, Message: "ok",
		Data: model.AppActionResponse{Success: true, Message: msg},
	})
}

func (s *Server) handleDockerImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	if !s.docker.Enabled() {
		writeJSON(w, http.StatusOK, model.APIResponse{
			Code: 0, Message: "ok",
			Data: model.ImageListResponse{Items: []model.Image{}},
		})
		return
	}

	images, err := s.docker.ListImages(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.ImageListResponse{Items: images},
	})
}

func (s *Server) handleProcessServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	services := s.process.ListServices(r.Context())
	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.ServiceListResponse{Items: services},
	})
}

func (s *Server) handleDiscoveredProcesses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	processes, err := s.process.DiscoverListeningProcesses(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.DiscoveredProcessListResponse{Items: processes},
	})
}

func (s *Server) handleTriggerScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	processes, err := s.process.TriggerScan(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.DiscoveredProcessListResponse{Items: processes},
	})
}

func (s *Server) handleDiscoveredAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/process/discovered/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "not found"})
		return
	}

	pid, err := strconv.Atoi(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "invalid pid"})
		return
	}

	action := parts[1]
	if action != "watch" {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "unsupported action: " + action})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	ok, msg, err := s.process.WatchDiscoveredApp(r.Context(), pid)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "app not found"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.AppActionResponse{Success: true, Message: msg},
	})
}

func (s *Server) handleProcessServiceAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/process/services/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 1 {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "not found"})
		return
	}

	name := parts[0]

	if len(parts) < 2 {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "action required"})
		return
	}

	action := parts[1]

	svcDef, found := s.process.FindService(name)
	if !found {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "service not found: " + name})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	var err error
	var msg string
	switch action {
	case "start":
		err = s.process.StartService(r.Context(), svcDef.Unit)
		msg = "已启动"
	case "stop":
		err = s.process.StopService(r.Context(), svcDef.Unit)
		msg = "已停止"
	case "restart":
		err = s.process.RestartService(r.Context(), svcDef.Unit)
		msg = "已重启"
	default:
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "unsupported action: " + action})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	operator := operatorFromContext(r.Context())
	s.audit.Record(operator, action, name, "process", msg)

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code: 0, Message: "ok",
		Data: model.AppActionResponse{Success: true, Message: msg},
	})
}

func (s *Server) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	logs := s.audit.List(limit, offset)
	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.OperationLogResponse{Items: logs},
	})
}

func (s *Server) handleWSDockerLogs(w http.ResponseWriter, r *http.Request) {
	if !s.docker.Enabled() {
		http.Error(w, "docker not available", http.StatusServiceUnavailable)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/ws/docker/logs/")
	if id == "" {
		http.Error(w, "container id required", http.StatusBadRequest)
		return
	}

	tail := 100
	if v := r.URL.Query().Get("tail"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tail = n
		}
	}

	ws.StreamDockerLogs(w, r, func(ctx context.Context, t int) (io.ReadCloser, error) {
		return s.docker.Logs(ctx, id, t, true)
	}, tail)
}

func (s *Server) handleWSProcessLogs(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/ws/process/logs/")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	svcDef, found := s.process.FindService(name)
	if !found {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	tail := 100
	if v := r.URL.Query().Get("tail"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			tail = n
		}
	}

	if svcDef.LogPath != "" {
		ch, err := s.process.TailLogFile(r.Context(), svcDef.LogPath, tail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ws.StreamFromChannel(w, r, ch)
		return
	}

	ch, err := s.process.LogStream(r.Context(), svcDef.Unit, tail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws.StreamFromChannel(w, r, ch)
}

func operatorFromContext(ctx context.Context) string {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return "anonymous"
	}
	return claims.Name
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
