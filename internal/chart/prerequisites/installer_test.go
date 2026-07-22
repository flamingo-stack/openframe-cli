package prerequisites

import (
	"testing"
)

func TestNewInstaller(t *testing.T) {
	installer := NewInstaller()

	if installer == nil {
		t.Fatal("Expected installer to be created")
		return
	}

	if installer.checker == nil {
		t.Error("Expected installer to have a checker")
	}
}
