package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func readProcUser(procDir string) string {
	info, err := os.Stat(procDir)
	if err != nil {
		return ""
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d", stat.Uid)
}

func getClockTicks() uint64 {
	return uint64(syscall.Getpagesize())
}
