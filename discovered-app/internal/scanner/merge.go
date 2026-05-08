package scanner

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Skyinfi/management-platform/discovered-app/internal/model"
)

func Merge(apps []*model.DiscoveredApp, pidPorts map[int][]int) []*model.DiscoveredApp {
	seen := make(map[int]bool)
	var result []*model.DiscoveredApp

	for _, app := range apps {
		if IsSystemProcess(app) {
			continue
		}

		if seen[app.PID] {
			continue
		}
		seen[app.PID] = true

		if ports, ok := pidPorts[app.PID]; ok {
			app.Ports = ports
		}

		Fingerprint(app)

		if app.Name == "" || app.Name == filepath.Base(app.ExePath) {
			if inferred := InferNameFromCWD(app.WorkDir); inferred != "" {
				app.Name = inferred
			}
		}

		app.InDocker = IsDockerProcess(app)
		app.Managed = IsManagedProcess(app)

		result = append(result, app)
	}
	return result
}

var systemProcessPrefixes = []string{
	"kworker",
	"ksoftirqd",
	"migration",
	"rcu_",
	"kcmed",
	"kthread",
	"kswapd",
	"jbd2",
	"ext4-rsv",
	"scsi_",
	"ata_",
	"md_",
	"edac-",
	"devfreq_wq",
	"pm_wq",
	"watchdog",
	"khelper",
	"kdevtmpfs",
	"netns",
}

var systemProcessNames = map[string]bool{
	"systemd":           true,
	"systemd-journal":   true,
	"systemd-udevd":     true,
	"systemd-resolve":   true,
	"systemd-timesyn":   true,
	"systemd-logind":    true,
	"systemd-network":   true,
	"sshd":              true,
	"cron":              true,
	"dbus-daemon":       true,
	"agetty":            true,
	"snapd":             true,
	"unattended-upgr":   true,
	"rsyslogd":          true,
	"irqbalance":        true,
	"polkitd":           true,
	"accounts-daemon":   true,
	"networkd-dispat":   true,
	"udisksd":           true,
	"ModemManager":      true,
	"powerd":            true,
	"thermald":          true,
	"fwupd":             true,
	"packagekitd":       true,
	"gpg-agent":         true,
	"dconf-service":     true,
	"gvfsd":             true,
	"pulseaudio":        true,
	"rtkit-daemon":      true,
	"colord":            true,
	"upowerd":           true,
	"boltd":             true,
	"udiskie":           true,
	"lvmetad":           true,
	"lvmlockd":          true,
	"dmeventd":          true,
	"auditd":            true,
	"rngd":              true,
	"chronyd":           true,
	"ntpd":              true,
	"haveged":           true,
	"multipathd":        true,
	"rpcbind":           true,
	"rpc.statd":         true,
	"rpc.idmapd":        true,
	"nfsd":              true,
	"lockd":             true,
	"mount.nfs":         true,
	"sm-notify":         true,
	"gssproxy":          true,
	"sssd":              true,
	"oddjobd":           true,
	"abrtd":             true,
	"abrt-watch-log":    true,
	"abrtd-dump-oops":   true,
	"ksmtuned":          true,
	"ksm":               true,
	"libvirtd":          true,
	"virtlogd":          true,
	"virtlockd":         true,
	"dnsmasq":           true,
	"dockerd":           true,
	"containerd":        true,
	"containerd-shim":   true,
	"docker-proxy":      true,
	"kubelet":           true,
	"kube-proxy":        true,
	"etcd":              true,
	"discovered-app":    true,
}

func IsSystemProcess(app *model.DiscoveredApp) bool {
	exe := filepath.Base(app.ExePath)
	cmdLine := strings.ToLower(app.CmdLine)

	if systemProcessNames[exe] {
		return true
	}

	firstArg := ""
	if parts := strings.Fields(cmdLine); len(parts) > 0 {
		firstArg = filepath.Base(parts[0])
	}
	if systemProcessNames[firstArg] {
		return true
	}

	for _, prefix := range systemProcessPrefixes {
		if strings.HasPrefix(exe, prefix) {
			return true
		}
	}

	if strings.Contains(cmdLine, "/usr/lib/systemd/") {
		return true
	}
	if strings.Contains(cmdLine, "/lib/systemd/") {
		return true
	}

	return false
}

func getProcRoot() string {
	if v := os.Getenv("PROC_ROOT"); v != "" {
		return v
	}
	return "/proc"
}

func IsDockerProcess(app *model.DiscoveredApp) bool {
	data, err := os.ReadFile(filepath.Join(getProcRoot(), strconv.Itoa(app.PID), "cgroup"))
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "docker") ||
		strings.Contains(content, "containerd") ||
		strings.Contains(content, "kubepods")
}

func IsManagedProcess(app *model.DiscoveredApp) bool {
	data, err := os.ReadFile(filepath.Join(getProcRoot(), strconv.Itoa(app.PID), "cgroup"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, ".service") {
			return true
		}
	}
	return false
}
