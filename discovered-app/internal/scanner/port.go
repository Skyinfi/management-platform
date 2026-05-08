package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseNetTCP(path string) (map[uint64]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	inodePorts := make(map[uint64]int)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		if fields[3] != "0A" {
			continue
		}

		localAddr := fields[1]
		colonIdx := strings.LastIndex(localAddr, ":")
		if colonIdx < 0 {
			continue
		}
		portHex := localAddr[colonIdx+1:]
		port, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			continue
		}

		inodeStr := fields[9]
		inode, err := strconv.ParseUint(inodeStr, 10, 64)
		if err != nil {
			continue
		}

		inodePorts[inode] = int(port)
	}
	return inodePorts, nil
}

func BuildPIDPortMap(procRoot string, inodePorts map[uint64]int) map[int][]int {
	pidPorts := make(map[int][]int)

	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return pidPorts
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join(procRoot, entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		seen := make(map[int]bool)
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}

			inode := parseSocketInode(link)
			if inode == 0 {
				continue
			}

			if port, ok := inodePorts[inode]; ok && !seen[port] {
				seen[port] = true
				pidPorts[pid] = append(pidPorts[pid], port)
			}
		}
	}
	return pidPorts
}

func parseSocketInode(link string) uint64 {
	if !strings.HasPrefix(link, "socket:[") {
		return 0
	}
	inodeStr := strings.TrimPrefix(link, "socket:[")
	inodeStr = strings.TrimSuffix(inodeStr, "]")
	inode, err := strconv.ParseUint(inodeStr, 10, 64)
	if err != nil {
		return 0
	}
	return inode
}

func LoadAllInodePorts(procRoot string) map[uint64]int {
	merged := make(map[uint64]int)

	for _, name := range []string{"net/tcp", "net/tcp6"} {
		path := filepath.Join(procRoot, name)
		m, err := ParseNetTCP(path)
		if err != nil {
			continue
		}
		for inode, port := range m {
			merged[inode] = port
		}
	}
	return merged
}
