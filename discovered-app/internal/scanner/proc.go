package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Skyinfi/management-platform/discovered-app/internal/model"
)

func ScanProc(procRoot string) ([]*model.DiscoveredApp, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, fmt.Errorf("read proc root %s: %w", procRoot, err)
	}

	var apps []*model.DiscoveredApp
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		app := readProcessInfo(procRoot, pid)
		if app != nil {
			apps = append(apps, app)
		}
	}
	return apps, nil
}

func readProcessInfo(procRoot string, pid int) *model.DiscoveredApp {
	procDir := filepath.Join(procRoot, strconv.Itoa(pid))

	cmdline := readProcFile(procDir, "cmdline")
	if cmdline == "" {
		return nil
	}

	exePath := readProcLink(procDir, "exe")
	cwd := readProcLink(procDir, "cwd")
	user := readProcUser(procDir)
	startTime := readProcStartTime(procRoot, pid)

	return &model.DiscoveredApp{
		PID:       pid,
		CmdLine:   cmdline,
		ExePath:   exePath,
		WorkDir:   cwd,
		User:      user,
		StartTime: startTime,
		Status:    "running",
		Ports:     []int{},
	}
}

func readProcFile(procDir, name string) string {
	data, err := os.ReadFile(filepath.Join(procDir, name))
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	return strings.Join(parts, " ")
}

func readProcLink(procDir, name string) string {
	value, err := os.Readlink(filepath.Join(procDir, name))
	if err != nil {
		return ""
	}
	return value
}

func readProcStartTime(procRoot string, pid int) time.Time {
	statData, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(pid), "stat"))
	if err != nil {
		return time.Time{}
	}

	fields := strings.Fields(string(statData))
	if len(fields) < 22 {
		return time.Time{}
	}

	comm := fields[1]
	if strings.HasPrefix(comm, "(") && strings.Contains(comm, ")") {
	} else if len(fields) > 22 {
		fields = fields[1:]
	}

	startTicks, err := strconv.ParseUint(fields[21], 10, 64)
	if err != nil {
		return time.Time{}
	}

	clockTicks := uint64(100)
	bootTime := getBootTime(procRoot)

	seconds := startTicks / clockTicks
	nanos := (startTicks % clockTicks) * 1000000000 / clockTicks

	t := bootTime.Add(time.Duration(seconds)*time.Second + time.Duration(nanos)*time.Nanosecond)
	return t
}

func getBootTime(procRoot string) time.Time {
	data, err := os.ReadFile(filepath.Join(procRoot, "stat"))
	if err != nil {
		return time.Now()
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "btime ") {
			continue
		}
		sec, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "btime ")), 10, 64)
		if err != nil {
			continue
		}
		return time.Unix(sec, 0)
	}
	return time.Now()
}
