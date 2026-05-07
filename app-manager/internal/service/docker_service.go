package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Skyinfi/management-platform/app-manager/internal/config"
	"github.com/Skyinfi/management-platform/app-manager/internal/model"

	containertypes "github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerService struct {
	cli    *client.Client
	mu     sync.RWMutex
	enabled bool
}

func NewDockerService(cfg config.DockerConfig) *DockerService {
	if !cfg.Enabled {
		return &DockerService{enabled: false}
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHost(cfg.Host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Printf("warning: docker client init failed: %v, docker features disabled", err)
		return &DockerService{enabled: false}
	}

	return &DockerService{cli: cli, enabled: true}
}

func (d *DockerService) Enabled() bool {
	return d.enabled
}

func (d *DockerService) ListContainers(ctx context.Context) ([]model.Container, error) {
	if !d.enabled {
		return nil, fmt.Errorf("docker not available")
	}

	containers, err := d.cli.ContainerList(ctx, containertypes.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]model.Container, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := make([]model.PortMapping, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, model.PortMapping{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}

		cont := model.Container{
			ID:        c.ID[:12],
			Name:      name,
			Image:     c.Image,
			Status:    c.Status,
			State:     c.State,
			Ports:     ports,
			Labels:    c.Labels,
			CreatedAt: c.Created,
		}

		if c.State == "running" {
			cpu, mem, _ := d.stats(ctx, c.ID)
			cont.CPU = cpu
			cont.Memory = mem
		}

		result = append(result, cont)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (d *DockerService) stats(ctx context.Context, containerID string) (cpu, memory string, err error) {
	stream, err := d.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return "—", "—", err
	}
	defer stream.Body.Close()

	data, err := io.ReadAll(stream.Body)
	if err != nil {
		return "—", "—", err
	}

	var stats struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs     int    `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	if err := parseJSON(data, &stats); err != nil {
		return "—", "—", err
	}

	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
	if sysDelta > 0 && stats.CPUStats.OnlineCPUs > 0 {
		cpuPct := (cpuDelta / sysDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		cpu = fmt.Sprintf("%.1f%%", cpuPct)
	} else {
		cpu = "—"
	}

	if stats.MemoryStats.Limit > 0 {
		memMB := float64(stats.MemoryStats.Usage) / 1024 / 1024
		memory = fmt.Sprintf("%.0f MB", memMB)
	} else {
		memory = "—"
	}

	return cpu, memory, nil
}

func (d *DockerService) StartContainer(ctx context.Context, id string) error {
	if !d.enabled {
		return fmt.Errorf("docker not available")
	}
	return d.cli.ContainerStart(ctx, id, containertypes.StartOptions{})
}

func (d *DockerService) StopContainer(ctx context.Context, id string) error {
	if !d.enabled {
		return fmt.Errorf("docker not available")
	}
	timeout := 30
	return d.cli.ContainerStop(ctx, id, containertypes.StopOptions{Timeout: &timeout})
}

func (d *DockerService) RestartContainer(ctx context.Context, id string) error {
	if !d.enabled {
		return fmt.Errorf("docker not available")
	}
	timeout := 30
	return d.cli.ContainerRestart(ctx, id, containertypes.StopOptions{Timeout: &timeout})
}

func (d *DockerService) RemoveContainer(ctx context.Context, id string) error {
	if !d.enabled {
		return fmt.Errorf("docker not available")
	}
	return d.cli.ContainerRemove(ctx, id, containertypes.RemoveOptions{Force: true})
}

func (d *DockerService) Logs(ctx context.Context, id string, tail int, follow bool) (io.ReadCloser, error) {
	if !d.enabled {
		return nil, fmt.Errorf("docker not available")
	}

	if tail <= 0 {
		tail = 100
	}

	return d.cli.ContainerLogs(ctx, id, containertypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       fmt.Sprintf("%d", tail),
		Timestamps: true,
	})
}

func (d *DockerService) ListImages(ctx context.Context) ([]model.Image, error) {
	if !d.enabled {
		return nil, fmt.Errorf("docker not available")
	}

	images, err := d.cli.ImageList(ctx, imagetypes.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	result := make([]model.Image, 0, len(images))
	for _, img := range images {
		repo, tag := "<none>", "<none>"
		if len(img.RepoTags) > 0 {
			parts := strings.SplitN(img.RepoTags[0], ":", 2)
			repo = parts[0]
			if len(parts) > 1 {
				tag = parts[1]
			}
		}

		result = append(result, model.Image{
			ID:         img.ID,
			Repository: repo,
			Tag:        tag,
			Size:       img.Size,
			CreatedAt:  img.Created,
			Names:      img.RepoTags,
		})
	}

	return result, nil
}

func (d *DockerService) Applications(ctx context.Context) []model.Application {
	containers, err := d.ListContainers(ctx)
	if err != nil {
		log.Printf("docker list containers: %v", err)
		return nil
	}

	apps := make([]model.Application, 0, len(containers))
	for _, c := range containers {
		status := c.State
		if status == "running" {
			status = "running"
		} else if status == "paused" {
			status = "warning"
		} else {
			status = "stopped"
		}

		endpoint := ""
		if len(c.Ports) > 0 && c.Ports[0].PublicPort > 0 {
			endpoint = fmt.Sprintf("0.0.0.0:%d", c.Ports[0].PublicPort)
		}

		owner := ""
		if c.Labels != nil {
			if v, ok := c.Labels["owner"]; ok {
				owner = v
			}
		}

		apps = append(apps, model.Application{
			Name:      c.Name,
			Type:      "Docker",
			Status:    status,
			Version:   "",
			Endpoint:  endpoint,
			Owner:     owner,
			CPU:       c.CPU,
			Memory:    c.Memory,
			Uptime:    formatUptime(c.CreatedAt),
			UpdatedAt: time.Unix(c.CreatedAt, 0).Format("15:04:05"),
		})
	}

	return apps
}

func formatUptime(created int64) string {
	d := time.Since(time.Unix(created, 0))
	if d < time.Minute {
		return "刚刚启动"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d 分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d 小时 %d 分钟", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%d 天 %d 小时", days, hours)
}

func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
