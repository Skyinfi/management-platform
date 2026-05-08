package scanner

import (
	"path/filepath"
	"strings"

	"github.com/Skyinfi/management-platform/discovered-app/internal/model"
)

func Fingerprint(app *model.DiscoveredApp) {
	exe := filepath.Base(app.ExePath)
	cmdLine := strings.ToLower(app.CmdLine)

	switch {
	case exe == "java" || strings.Contains(cmdLine, ".jar"):
		app.Type = "java"
		app.Name = extractJavaAppName(app.CmdLine)
	case exe == "python" || exe == "python3" || strings.HasPrefix(exe, "python"):
		app.Type = "python"
		app.Name = extractPythonAppName(app.CmdLine)
	case exe == "node" || strings.HasSuffix(exe, "node"):
		app.Type = "node"
		app.Name = extractNodeAppName(app.CmdLine)
	default:
		app.Type = "go"
	}

	if app.Name == "" {
		app.Name = exe
	}
}

func extractJavaAppName(cmdLine string) string {
	parts := strings.Fields(cmdLine)
	for i, p := range parts {
		if p == "-jar" && i+1 < len(parts) {
			jar := parts[i+1]
			jar = filepath.Base(jar)
			return strings.TrimSuffix(jar, ".jar")
		}
	}
	return ""
}

func extractPythonAppName(cmdLine string) string {
	parts := strings.Fields(cmdLine)
	for _, p := range parts {
		if strings.HasSuffix(p, ".py") {
			return strings.TrimSuffix(filepath.Base(p), ".py")
		}
	}
	return ""
}

func extractNodeAppName(cmdLine string) string {
	parts := strings.Fields(cmdLine)
	for _, p := range parts {
		if strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".mjs") {
			return strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		}
	}
	return ""
}
