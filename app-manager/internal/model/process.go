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
