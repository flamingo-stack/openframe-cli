package memory

import (
	"fmt"
	"math"

	sysinfo "github.com/elastic/go-sysinfo"
)

type MemoryChecker struct{}

const RecommendedMemoryMB = 15360 // 15GB in MB

func NewMemoryChecker() *MemoryChecker {
	return &MemoryChecker{}
}

func (m *MemoryChecker) IsInstalled() bool {
	return m.HasSufficientMemory()
}

func (m *MemoryChecker) GetInstallHelp() string {
	currentMemory := m.getTotalMemoryMB()
	return fmt.Sprintf("Memory: %d MB available, %d MB recommended. Consider adding more RAM or increasing virtual memory", currentMemory, RecommendedMemoryMB)
}

func (m *MemoryChecker) Install() error {
	return fmt.Errorf("memory cannot be automatically installed. Please add more physical RAM or increase virtual memory allocation")
}

func (m *MemoryChecker) HasSufficientMemory() bool {
	totalMemory := m.getTotalMemoryMB()
	return totalMemory >= RecommendedMemoryMB
}

// getTotalMemoryMB returns total physical RAM in MB, read cross-platform via
// elastic/go-sysinfo (procfs on Linux, sysctl syscall on macOS, Win32 API on
// Windows — no sysctl/cat/powershell shell-outs). Returns 0 if unavailable.
func (m *MemoryChecker) getTotalMemoryMB() int {
	host, err := sysinfo.Host()
	if err != nil {
		return 0
	}
	memInfo, err := host.Memory()
	if err != nil {
		return 0
	}
	mb := memInfo.Total / (1024 * 1024)
	if mb > uint64(math.MaxInt) {
		return math.MaxInt
	}
	return int(mb)
}

func (m *MemoryChecker) GetMemoryInfo() (int, int, bool) {
	current := m.getTotalMemoryMB()
	recommended := RecommendedMemoryMB
	sufficient := current >= recommended
	return current, recommended, sufficient
}
