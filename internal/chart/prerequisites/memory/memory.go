package memory

import (
	"fmt"
	"math"

	sysmem "github.com/pbnjay/memory"
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

// getTotalMemoryMB returns total physical RAM in MB, read cross-platform via a
// syscall (no sysctl/cat/powershell shell-outs). Returns 0 if unavailable.
func (m *MemoryChecker) getTotalMemoryMB() int {
	mb := sysmem.TotalMemory() / (1024 * 1024)
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
