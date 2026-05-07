package model

type Application struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Version   string `json:"version"`
	Endpoint  string `json:"endpoint"`
	Owner     string `json:"owner"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	Uptime    string `json:"uptime"`
	UpdatedAt string `json:"updatedAt"`
}

type ActivityItem struct {
	Time    string `json:"time"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Tone    string `json:"tone"`
	AppName string `json:"appName"`
}

type MetricItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
}

type DashboardResponse struct {
	Metrics      []MetricItem   `json:"metrics"`
	Activities   []ActivityItem `json:"activities"`
	Applications []Application  `json:"applications"`
}

type ListApplicationsResponse struct {
	Items []Application `json:"items"`
}

type AppActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ApplicationLogResponse struct {
	Lines []string `json:"lines"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
