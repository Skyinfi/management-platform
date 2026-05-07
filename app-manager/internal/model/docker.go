package model

type Container struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Status    string            `json:"status"`
	State     string            `json:"state"`
	Ports     []PortMapping     `json:"ports"`
	CPU       string            `json:"cpu"`
	Memory    string            `json:"memory"`
	Labels    map[string]string `json:"labels"`
	CreatedAt int64             `json:"createdAt"`
}

type PortMapping struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"privatePort"`
	PublicPort  uint16 `json:"publicPort"`
	Type        string `json:"type"`
}

type Image struct {
	ID         string   `json:"id"`
	Repository string   `json:"repository"`
	Tag        string   `json:"tag"`
	Size       int64    `json:"size"`
	CreatedAt  int64    `json:"createdAt"`
	Names      []string `json:"names"`
}

type ContainerListResponse struct {
	Items []Container `json:"items"`
}

type ImageListResponse struct {
	Items []Image `json:"items"`
}
