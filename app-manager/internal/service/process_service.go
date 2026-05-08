package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Skyinfi/management-platform/app-manager/internal/config"
	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

type ProcessService struct {
	services []config.ServiceDef
	scanner  *ScannerClient
	mu       sync.RWMutex
}

func NewProcessService(services []config.ServiceDef, scanner *ScannerClient) *ProcessService {
	return &ProcessService{services: services, scanner: scanner}
}

func (p *ProcessService) ListServices(ctx context.Context) []model.ManagedService {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]model.ManagedService, 0, len(p.services))
	for _, svc := range p.services {
		ms := model.ManagedService{
			Name:     svc.Name,
			Display:  svc.Display,
			Type:     svc.Type,
			Unit:     svc.Unit,
			Endpoint: svc.Endpoint,
			Owner:    svc.Owner,
		}

		status, pid, uptime := p.queryUnitStatus(ctx, svc.Unit)
		ms.Status = status
		ms.PID = pid
		ms.Uptime = uptime

		result = append(result, ms)
	}
	return result
}

func (p *ProcessService) queryUnitStatus(ctx context.Context, unit string) (status string, pid int, uptime string) {
	out, err := exec.CommandContext(ctx, "systemctl", "show", unit,
		"--property=ActiveState,MainPID,ActiveEnterTimestamp").CombinedOutput()
	if err != nil {
		return "unknown", 0, ""
	}

	activeState := ""
	activeEnter := ""
	mainPID := 0

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ActiveState=") {
			activeState = strings.TrimPrefix(line, "ActiveState=")
		} else if strings.HasPrefix(line, "MainPID=") {
			val := strings.TrimPrefix(line, "MainPID=")
			mainPID, _ = strconv.Atoi(val)
		} else if strings.HasPrefix(line, "ActiveEnterTimestamp=") {
			activeEnter = strings.TrimPrefix(line, "ActiveEnterTimestamp=")
		}
	}

	switch activeState {
	case "active":
		status = "running"
	case "inactive":
		status = "stopped"
	case "failed":
		status = "error"
	case "activating":
		status = "starting"
	default:
		status = activeState
	}

	pid = mainPID

	if activeEnter != "" && activeEnter != "n/a" {
		t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", activeEnter)
		if err == nil {
			d := time.Since(t)
			uptime = formatDuration(d)
		}
	}

	return status, pid, uptime
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d 秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d 分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d 小时 %d 分", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%d 天 %d 小时", days, hours)
}

func (p *ProcessService) StartService(ctx context.Context, unit string) error {
	out, err := exec.CommandContext(ctx, "systemctl", "start", unit).CombinedOutput()
	if err != nil {
		return fmt.Errorf("start %s: %s: %w", unit, string(out), err)
	}
	return nil
}

func (p *ProcessService) StopService(ctx context.Context, unit string) error {
	out, err := exec.CommandContext(ctx, "systemctl", "stop", unit).CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop %s: %s: %w", unit, string(out), err)
	}
	return nil
}

func (p *ProcessService) RestartService(ctx context.Context, unit string) error {
	out, err := exec.CommandContext(ctx, "systemctl", "restart", unit).CombinedOutput()
	if err != nil {
		return fmt.Errorf("restart %s: %s: %w", unit, string(out), err)
	}
	return nil
}

func (p *ProcessService) Logs(ctx context.Context, unit string, tail int) ([]string, error) {
	if tail <= 0 {
		tail = 100
	}

	out, err := exec.CommandContext(ctx, "journalctl", "-u", unit,
		fmt.Sprintf("-n%d", tail), "--no-pager").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("journalctl %s: %w", unit, err)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	return lines, nil
}

func (p *ProcessService) LogStream(ctx context.Context, unit string, tail int) (<-chan string, error) {
	if tail <= 0 {
		tail = 100
	}

	cmd := exec.CommandContext(ctx, "journalctl", "-u", unit,
		fmt.Sprintf("-n%d", tail), "--no-pager", "-f")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("journalctl pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("journalctl start: %w", err)
	}

	ch := make(chan string, 256)
	go func() {
		defer close(ch)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case ch <- scanner.Text():
			case <-ctx.Done():
				cmd.Process.Kill()
				return
			}
		}
	}()

	return ch, nil
}

func (p *ProcessService) TailLogFile(ctx context.Context, logPath string, tail int) (<-chan string, error) {
	if tail <= 0 {
		tail = 100
	}

	cmd := exec.CommandContext(ctx, "tail", "-n", strconv.Itoa(tail), "-f", logPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("tail pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("tail start: %w", err)
	}

	ch := make(chan string, 256)
	go func() {
		defer close(ch)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			select {
			case ch <- scanner.Text():
			case <-ctx.Done():
				cmd.Process.Kill()
				return
			}
		}
	}()

	return ch, nil
}

func (p *ProcessService) LogReader(ctx context.Context, name string, tail int) (io.ReadCloser, error) {
	for _, svc := range p.services {
		if svc.Name != name && svc.Unit != name {
			continue
		}
		if svc.LogPath != "" {
			args := []string{"-n", strconv.Itoa(tail)}
			cmd := exec.CommandContext(ctx, "cat", svc.LogPath)
			cmd.Args = append(args, svc.LogPath)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return nil, err
			}
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			return stdout, nil
		}
	}

	lines, err := p.Logs(ctx, name, tail)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(strings.Join(lines, "\n"))), nil
}

func (p *ProcessService) FindService(name string) (config.ServiceDef, bool) {
	for _, svc := range p.services {
		if svc.Name == name || svc.Unit == name {
			return svc, true
		}
	}
	return config.ServiceDef{}, false
}

func (p *ProcessService) DiscoverListeningProcesses(ctx context.Context) ([]model.DiscoveredProcess, error) {
	if p.scanner == nil {
		return []model.DiscoveredProcess{}, nil
	}
	apps, err := p.scanner.GetApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanner get apps: %w", err)
	}
	log.Printf("discover listening processes via scanner agent completed, total=%d", len(apps))
	return apps, nil
}

func (p *ProcessService) TriggerScan(ctx context.Context) ([]model.DiscoveredProcess, error) {
	if p.scanner == nil {
		return nil, fmt.Errorf("scanner agent not configured")
	}
	apps, err := p.scanner.RunScan(ctx)
	if err != nil {
		return nil, fmt.Errorf("scanner run scan: %w", err)
	}
	log.Printf("triggered scan via scanner agent, discovered=%d", len(apps))
	return apps, nil
}

func (p *ProcessService) WatchDiscoveredApp(ctx context.Context, pid int) (bool, string, error) {
	if p.scanner == nil {
		return false, "", fmt.Errorf("scanner agent not configured")
	}
	return p.scanner.WatchApp(ctx, pid)
}

func (p *ProcessService) Applications(ctx context.Context) []model.Application {
	services := p.ListServices(ctx)
	apps := make([]model.Application, 0, len(services))
	for _, svc := range services {
		status := svc.Status
		if status == "error" {
			status = "warning"
		}

		cpu := "—"
		mem := "—"
		if svc.Status == "running" && svc.PID > 0 {
			cpu, mem = getProcessStats(svc.PID)
		}

		apps = append(apps, model.Application{
			Name:      svc.Name,
			Type:      "Process",
			Status:    status,
			Version:   "",
			Endpoint:  svc.Endpoint,
			Owner:     svc.Owner,
			CPU:       cpu,
			Memory:    mem,
			Uptime:    svc.Uptime,
			UpdatedAt: "",
		})
	}
	return apps
}

func getProcessStats(pid int) (cpu, memory string) {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "%cpu,rss", "--no-headers").Output()
	if err != nil {
		return "—", "—"
	}

	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return "—", "—"
	}

	cpu = strings.TrimSpace(fields[0]) + "%"
	rssKB, _ := strconv.ParseFloat(fields[1], 64)
	memory = fmt.Sprintf("%.0f MB", rssKB/1024)

	return cpu, memory
}
