package scanner

import "os"

func readProcUser(procDir string) string {
	info, err := os.Stat(procDir)
	if err != nil {
		return ""
	}
	_ = info
	return "0"
}

func getClockTicks() uint64 {
	return 100
}
