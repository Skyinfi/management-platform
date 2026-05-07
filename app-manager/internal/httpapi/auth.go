package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{Code: 400, Message: "invalid json"})
		return
	}

	resp, err := s.auth.Login(req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.APIResponse{Code: 401, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "ok", Data: resp})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, model.APIResponse{Code: 405, Message: "method not allowed"})
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, model.APIResponse{Code: 401, Message: "missing authorization"})
		return
	}

	resp, err := s.auth.Me(token[len("Bearer "):])
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, model.APIResponse{Code: 401, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{Code: 0, Message: "ok", Data: resp})
}
