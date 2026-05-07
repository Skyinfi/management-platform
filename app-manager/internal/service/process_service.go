package service

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Skyinfi/management-platform/app-manager/internal/config"
	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

type ProcessService struct {
	services []config.ServiceDef
	mu       sync.RWMutex
}

var (
	ssPIDPattern  = regexp.MustCompile(`pid=(\d+)`)
	ssNamePattern = regexp.MustCompile(`"([^"]+)"`)
)

func NewProcessService(services []config.ServiceDef) *ProcessService {
	return &ProcessService{services: services}
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
	managedPIDs := map[int]bool{}
	for _, svc := range p.ListServices(ctx) {
		if svc.PID > 0 {
			managedPIDs[svc.PID] = true
		}
	}

	discovered := []model.DiscoveredProcess{}
	if _, err := exec.LookPath("ss"); err == nil {
		out, err := exec.CommandContext(ctx, "ss", "-H", "-ltnp").CombinedOutput()
		if err == nil {
			discovered = append(discovered, p.discoverFromSSOutput(string(out), managedPIDs)...)
		}
	}

	if _, err := exec.LookPath("sudo"); err == nil {
		out, err := exec.CommandContext(ctx, "sudo", "-n", "ss", "-H", "-ltnp").CombinedOutput()
		if err == nil {
			discovered = append(discovered, p.discoverFromSSOutput(string(out), managedPIDs)...)
		}
	}

	if _, err := exec.LookPath("lsof"); err == nil {
		out, err := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN").CombinedOutput()
		if err == nil {
			discovered = append(discovered, p.discoverFromLSOFOutput(string(out), managedPIDs)...)
		}
	}

	if _, err := exec.LookPath("sudo"); err == nil {
		out, err := exec.CommandContext(ctx, "sudo", "-n", "lsof", "-nP", "-iTCP", "-sTCP:LISTEN").CombinedOutput()
		if err == nil {
			discovered = append(discovered, p.discoverFromLSOFOutput(string(out), managedPIDs)...)
		}
	}

	return dedupeDiscoveredProcesses(discovered), nil
}

func (p *ProcessService) discoverFromSSOutput(output string, managedPIDs map[int]bool) []model.DiscoveredProcess {
	seen := map[string]bool{}
	discovered := []model.DiscoveredProcess{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		item, ok := p.discoveredFromSSLine(line, managedPIDs)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%d|%s", item.PID, item.Endpoint)
		if seen[key] {
			continue
		}
		seen[key] = true
		discovered = append(discovered, item)
	}

	return discovered
}

func (p *ProcessService) discoverFromLSOFOutput(output string, managedPIDs map[int]bool) []model.DiscoveredProcess {
	seen := map[string]bool{}
	discovered := []model.DiscoveredProcess{}
	for index, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || index == 0 {
			continue
		}

		item, ok := p.discoveredFromLSOFLine(line, managedPIDs)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%d|%s", item.PID, item.Endpoint)
		if seen[key] {
			continue
		}
		seen[key] = true
		discovered = append(discovered, item)
	}

	return discovered
}

func (p *ProcessService) discoveredFromSSLine(line string, managedPIDs map[int]bool) (model.DiscoveredProcess, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return model.DiscoveredProcess{}, false
	}

	endpoint := fields[3]
	pid := parseSSPID(line)
	if pid <= 0 {
		return model.DiscoveredProcess{
			ID:        discoveryID(0, endpoint),
			Name:      "unknown",
			Endpoint:  endpoint,
			Protocol:  "tcp",
			Command:   "权限不足，无法读取进程信息",
			Adoptable: false,
		}, true
	}

	name := parseSSProcessName(line)
	exePath := readProcLink(pid, "exe")
	cwd := readProcLink(pid, "cwd")
	command := readProcCmdline(pid)
	inDocker := processInDocker(pid)

	if name == "" {
		name = processNameFromCommand(command, exePath)
	}

	id := discoveryID(pid, endpoint)
	return model.DiscoveredProcess{
		ID:        id,
		Name:      name,
		PID:       pid,
		User:      processUser(pid),
		Endpoint:  endpoint,
		Protocol:  "tcp",
		Command:   command,
		ExePath:   exePath,
		Cwd:       cwd,
		Managed:   managedPIDs[pid],
		InDocker:  inDocker,
		Adoptable: isAdoptableProcess(managedPIDs[pid], inDocker, cwd, exePath, command),
	}, true
}

func (p *ProcessService) discoveredFromLSOFLine(line string, managedPIDs map[int]bool) (model.DiscoveredProcess, bool) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return model.DiscoveredProcess{}, false
	}

	pid, err := strconv.Atoi(fields[1])
	if err != nil || pid <= 0 {
		return model.DiscoveredProcess{}, false
	}

	name := fields[0]
	userName := fields[2]
	endpoint := strings.Join(fields[8:], " ")
	endpoint = strings.TrimSuffix(endpoint, " (LISTEN)")
	exePath := readProcLink(pid, "exe")
	cwd := readProcLink(pid, "cwd")
	command := readProcCmdline(pid)
	inDocker := processInDocker(pid)

	if command == "" {
		command = name
	}

	id := discoveryID(pid, endpoint)
	return model.DiscoveredProcess{
		ID:        id,
		Name:      name,
		PID:       pid,
		User:      userName,
		Endpoint:  endpoint,
		Protocol:  "tcp",
		Command:   command,
		ExePath:   exePath,
		Cwd:       cwd,
		Managed:   managedPIDs[pid],
		InDocker:  inDocker,
		Adoptable: isAdoptableProcess(managedPIDs[pid], inDocker, cwd, exePath, command),
	}, true
}

func dedupeDiscoveredProcesses(items []model.DiscoveredProcess) []model.DiscoveredProcess {
	seen := map[string]bool{}
	result := make([]model.DiscoveredProcess, 0, len(items))
	for _, item := range items {
		key := fmt.Sprintf("%d|%s", item.PID, item.Endpoint)
		if item.PID == 0 {
			key = item.Endpoint
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, item)
	}
	return result
}

func isAdoptableProcess(managed, inDocker bool, cwd, exePath, command string) bool {
	if managed || inDocker {
		return false
	}
	return hasOptPath(cwd) || hasOptPath(exePath) || hasOptPath(command)
}

func hasOptPath(value string) bool {
	return value == "/opt" ||
		strings.HasPrefix(value, "/opt/") ||
		strings.Contains(value, " /opt/") ||
		strings.Contains(value, "=/opt/")
}

func parseSSPID(line string) int {
	match := ssPIDPattern.FindStringSubmatch(line)
	if len(match) != 2 {
		return 0
	}
	pid, _ := strconv.Atoi(match[1])
	return pid
}

func parseSSProcessName(line string) string {
	match := ssNamePattern.FindStringSubmatch(line)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func readProcLink(pid int, name string) string {
	value, err := os.Readlink(fmt.Sprintf("/proc/%d/%s", pid, name))
	if err != nil {
		return ""
	}
	return value
}

func readProcCmdline(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	return strings.Join(parts, " ")
}

func processNameFromCommand(command, exePath string) string {
	if command != "" {
		fields := strings.Fields(command)
		if len(fields) > 0 {
			return fields[0]
		}
	}
	if exePath != "" {
		parts := strings.Split(strings.TrimRight(exePath, "/"), "/")
		return parts[len(parts)-1]
	}
	return "unknown"
}

func processInDocker(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return false
	}
	value := string(data)
	return strings.Contains(value, "docker") ||
		strings.Contains(value, "containerd") ||
		strings.Contains(value, "kubepods")
}

func processUser(pid int) string {
	info, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return ""
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	u, err := user.LookupId(uid)
	if err != nil {
		return uid
	}
	return u.Username
}

func discoveryID(pid int, endpoint string) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%d|%s", pid, endpoint)))
	return hex.EncodeToString(sum[:8])
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
