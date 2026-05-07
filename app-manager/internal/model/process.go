package model

type ManagedService struct {
	Name     string `json:"name"`
	Display  string `json:"display"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Unit     string `json:"unit"`
	Endpoint string `json:"endpoint"`
	Owner    string `json:"owner"`
	PID      int    `json:"pid"`
	Uptime   string `json:"uptime"`
}

type ServiceListResponse struct {
	Items []ManagedService `json:"items"`
}

type DiscoveredProcess struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PID       int    `json:"pid"`
	User      string `json:"user"`
	Endpoint  string `json:"endpoint"`
	Protocol  string `json:"protocol"`
	Command   string `json:"command"`
	ExePath   string `json:"exePath"`
	Cwd       string `json:"cwd"`
	Managed   bool   `json:"managed"`
	InDocker  bool   `json:"inDocker"`
	Adoptable bool   `json:"adoptable"`
}

type DiscoveredProcessListResponse struct {
	Items []DiscoveredProcess `json:"items"`
}
