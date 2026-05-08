package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Skyinfi/management-platform/discovered-app/internal/model"
	"github.com/Skyinfi/management-platform/discovered-app/internal/scanner"
	"github.com/Skyinfi/management-platform/discovered-app/internal/ws"
)

type Handler struct {
	agent *scanner.Agent
	mux   *http.ServeMux
}

func NewHandler(agent *scanner.Agent) *Handler {
	h := &Handler{
		agent: agent,
		mux:   http.NewServeMux(),
	}
	h.routes()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("/api/health", h.handleHealth)
	h.mux.HandleFunc("/api/scanner/run", h.handleScanRun)
	h.mux.HandleFunc("/api/scanner/apps", h.handleApps)
	h.mux.HandleFunc("/api/scanner/apps/", h.handleAppDetail)
	h.mux.HandleFunc("/ws/scanner/progress", h.handleWSProgress)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) handleScanRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	apps, err := h.agent.RunScan(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data: model.ScanResponse{
			Count: len(apps),
			Apps:  apps,
		},
	})
}

func (h *Handler) handleApps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	apps := h.agent.GetApps()
	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data: model.AppListResponse{
			Items: apps,
		},
	})
}

func (h *Handler) handleAppDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/scanner/apps/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "pid required"})
		return
	}

	parts := strings.Split(path, "/")
	pid, err := strconv.Atoi(parts[0])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "invalid pid"})
		return
	}

	if len(parts) >= 2 && parts[1] == "watch" {
		h.handleWatch(w, r, pid)
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	app := h.agent.GetApp(pid)
	if app == nil {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "app not found"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data: model.AppDetailResponse{
			App: app,
		},
	})
}

func (h *Handler) handleWatch(w http.ResponseWriter, r *http.Request, pid int) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	ok := h.agent.WatchApp(pid)
	if !ok {
		writeJSON(w, http.StatusNotFound, model.APIResponse{Code: 404, Message: "app not found"})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
		Code:    0,
		Message: "ok",
		Data: model.WatchResponse{
			Success: true,
			Message: "已纳入监控",
		},
	})
}

func (h *Handler) handleWSProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}
	ws.StreamProgress(w, r, h.agent.ProgressCh())
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
