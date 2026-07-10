package memory

import (
	"testing"
)

func TestNewMemoryChecker(t *testing.T) {
	checker := NewMemoryChecker()

	if checker == nil {
		t.Error("Expected Memory checker to be created")
	}
}

func TestMemoryChecker_GetMemoryInfo(t *testing.T) {
	checker := NewMemoryChecker()
	current, recommended, sufficient := checker.GetMemoryInfo()

	if current <= 0 {
		t.Error("Current memory should be greater than 0")
	}

	if recommended != RecommendedMemoryMB {
		t.Errorf("Expected recommended memory to be %d, got %d", RecommendedMemoryMB, recommended)
	}

	expectedSufficient := current >= recommended
	if sufficient != expectedSufficient {
		t.Errorf("Expected sufficient to be %v, got %v", expectedSufficient, sufficient)
	}
}

func TestMemoryChecker_GetInstallHelp(t *testing.T) {
	checker := NewMemoryChecker()
	help := checker.GetInstallHelp()

	if help == "" {
		t.Error("Install help should not be empty")
	}

	// Should contain memory information
	if !containsSubstring(help, "MB") {
		t.Errorf("Help should contain memory information in MB: %s", help)
	}

	if !containsSubstring(help, "recommended") {
		t.Errorf("Help should mention recommended memory: %s", help)
	}
}

func TestGetTotalMemoryMB(t *testing.T) {
	checker := NewMemoryChecker()
	mem := checker.getTotalMemoryMB()

	// go-sysinfo reads real physical RAM, so any machine running the test suite
	// must report a positive, sane amount (>= 256MB).
	if mem < 256 {
		t.Errorf("expected total memory >= 256MB, got %d MB", mem)
	}

	// Stable across calls (no shell-out flakiness).
	if again := checker.getTotalMemoryMB(); again != mem {
		t.Errorf("total memory must be stable across calls: %d vs %d", mem, again)
	}
}

func TestHasSufficientMemory(t *testing.T) {
	checker := NewMemoryChecker()

	// Test the logic
	sufficient := checker.HasSufficientMemory()
	totalMemory := checker.getTotalMemoryMB()
	expectedSufficient := totalMemory >= RecommendedMemoryMB

	if sufficient != expectedSufficient {
		t.Errorf("HasSufficientMemory() = %v, expected %v (total: %d MB, recommended: %d MB)",
			sufficient, expectedSufficient, totalMemory, RecommendedMemoryMB)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
