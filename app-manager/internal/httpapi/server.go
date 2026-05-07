package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Skyinfi/app-manager/internal/middleware"
	"github.com/Skyinfi/app-manager/internal/model"
	"github.com/Skyinfi/app-manager/internal/service"
)

type Server struct {
	service  *service.Service
	auth     *service.AuthService
	mux      *http.ServeMux
	security middleware.JWTValidator
	cors     bool
	origin   string
}

func New(svc *service.Service, auth *service.AuthService, opts ...Option) *Server {
	s := &Server{service: svc, auth: auth, mux: http.NewServeMux()}
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
		Data:    s.service.Dashboard(),
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
		Data:    model.ListApplicationsResponse{Items: s.service.Applications()},
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

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data:    model.AppActionResponse{Success: true, Message: updated},
	})
}



func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
