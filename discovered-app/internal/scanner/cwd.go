package scanner

import (
	"path/filepath"
	"strings"
)

func InferNameFromCWD(cwd string) string {
	if cwd == "" {
		return ""
	}

	excluded := map[string]bool{
		"/":     true,
		"/home": true,
		"/tmp":  true,
		"/root": true,
		"/var":  true,
		"/etc":  true,
		"/usr":  true,
		"/opt":  true,
		"/srv":  true,
	}

	if excluded[cwd] {
		return ""
	}

	base := filepath.Base(cwd)
	if base == "" || base == "." || base == "/" {
		return ""
	}

	lower := strings.ToLower(base)
	skip := map[string]bool{
		"bin": true, "sbin": true, "lib": true, "src": true,
		"config": true, "data": true, "log": true, "logs": true,
		"cache": true, "run": true, "tmp": true, "temp": true,
	}
	if skip[lower] {
		return ""
	}

	return base
}
