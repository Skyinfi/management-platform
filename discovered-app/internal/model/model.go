package model

import "time"

type DiscoveredApp struct {
	PID      int       `json:"pid"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	CmdLine  string    `json:"cmdLine"`
	WorkDir  string    `json:"workDir"`
	ExePath  string    `json:"exePath"`
	Ports    []int     `json:"ports"`
	User     string    `json:"user"`
	StartTime time.Time `json:"startTime"`
	Managed  bool      `json:"managed"`
	Status   string    `json:"status"`
	InDocker bool      `json:"inDocker"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ScanResponse struct {
	Count int             `json:"count"`
	Apps  []*DiscoveredApp `json:"apps"`
}

type AppListResponse struct {
	Items []*DiscoveredApp `json:"items"`
}

type AppDetailResponse struct {
	App *DiscoveredApp `json:"app"`
}

type WatchResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ProgressMessage struct {
	Phase   string `json:"phase"`
	Count   int    `json:"count"`
	Message string `json:"message"`
	Done    bool   `json:"done"`
}
