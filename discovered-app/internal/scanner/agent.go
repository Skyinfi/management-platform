package scanner

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Skyinfi/management-platform/discovered-app/internal/model"
)

type Agent struct {
	procRoot     string
	mu           sync.RWMutex
	apps         []*model.DiscoveredApp
	progressCh   chan model.ProgressMessage
	lastScanTime time.Time
}

func NewAgent(procRoot string) *Agent {
	return &Agent{
		procRoot:   procRoot,
		progressCh: make(chan model.ProgressMessage, 64),
	}
}

func (a *Agent) RunScan(ctx context.Context) ([]*model.DiscoveredApp, error) {
	var (
		procApps []*model.DiscoveredApp
		pidPorts map[int][]int
		procErr  error
		portErr  error
	)

	a.sendProgress("scan_start", 0, "开始扫描...")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		a.sendProgress("scanning_proc", 0, "扫描 /proc 文件系统...")
		apps, err := ScanProc(a.procRoot)
		if err != nil {
			procErr = err
			return
		}
		procApps = apps
		a.sendProgress("proc_done", len(apps), "发现进程")
	}()

	go func() {
		defer wg.Done()
		a.sendProgress("scanning_ports", 0, "扫描端口...")
		inodePorts := LoadAllInodePorts(a.procRoot)
		pidPorts = BuildPIDPortMap(a.procRoot, inodePorts)
		totalPorts := 0
		for _, ports := range pidPorts {
			totalPorts += len(ports)
		}
		a.sendProgress("port_done", totalPorts, "发现端口绑定")
	}()

	wg.Wait()

	if procErr != nil {
		return nil, procErr
	}
	if portErr != nil {
		log.Printf("port scan warning: %v", portErr)
	}

	a.sendProgress("merging", 0, "聚合结果、过滤系统进程...")
	merged := Merge(procApps, pidPorts)

	a.sendProgress("scan_complete", len(merged), "扫描完成")

	a.mu.Lock()
	a.apps = merged
	a.lastScanTime = time.Now()
	a.mu.Unlock()

	return merged, nil
}

func (a *Agent) GetApps() []*model.DiscoveredApp {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.apps
}

func (a *Agent) GetApp(pid int) *model.DiscoveredApp {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, app := range a.apps {
		if app.PID == pid {
			return app
		}
	}
	return nil
}

func (a *Agent) WatchApp(pid int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, app := range a.apps {
		if app.PID == pid {
			app.Managed = true
			return true
		}
	}
	return false
}

func (a *Agent) LastScanTime() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastScanTime
}

func (a *Agent) StartScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	go func() {
		_, _ = a.RunScan(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := a.RunScan(ctx)
			if err != nil {
				log.Printf("scheduled scan failed: %v", err)
			}
		}
	}
}

func (a *Agent) ProgressCh() <-chan model.ProgressMessage {
	return a.progressCh
}

func (a *Agent) sendProgress(phase string, count int, message string) {
	select {
	case a.progressCh <- model.ProgressMessage{
		Phase:   phase,
		Count:   count,
		Message: message,
		Done:    phase == "scan_complete",
	}:
	default:
	}
}
