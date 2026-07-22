package prerequisites

import (
	"runtime"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/docker"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/helm"
	"github.com/flamingo-stack/openframe-cli/internal/cluster/prerequisites/k3d"
)

func TestNewPrerequisiteChecker(t *testing.T) {
	checker := NewPrerequisiteChecker()

	if len(checker.requirements) != 3 {
		t.Errorf("Expected 3 requirements, got %d", len(checker.requirements))
	}

	// kubectl is intentionally absent — the CLI uses client-go, not the binary.
	expectedNames := []string{"Docker", "k3d", "helm"}
	for i, req := range checker.requirements {
		if req.Name != expectedNames[i] {
			t.Errorf("Expected requirement %d to be %s, got %s", i, expectedNames[i], req.Name)
		}
	}
}

func TestCommandExists(t *testing.T) {
	// Test using docker package since it has commandExists function
	dockerInstaller := docker.NewDockerInstaller()

	// We can't directly test commandExists since it's not exported,
	// but we can test IsInstalled which uses it internally
	_ = dockerInstaller.IsInstalled()
}

func TestInstallHelp(t *testing.T) {
	tests := []struct {
		name     string
		helpFunc func() string
	}{
		{"docker", docker.NewDockerInstaller().GetInstallHelp},
		{"k3d", k3d.NewK3dInstaller().GetInstallHelp},
		{"helm", helm.NewHelmInstaller().GetInstallHelp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.helpFunc()
			if help == "" {
				t.Errorf("Install help for %s should not be empty", tt.name)
			}

			switch runtime.GOOS {
			case "darwin":
				if !containsAny(help, []string{"brew", "https://"}) {
					t.Errorf("macOS help should contain brew or https reference: %s", help)
				}
			case "linux":
				if !containsAny(help, []string{"package manager", "https://", "curl"}) {
					t.Errorf("Linux help should contain package manager, https, or curl reference: %s", help)
				}
			case "windows":
				if !containsAny(help, []string{"https://", "chocolatey", "choco"}) {
					t.Errorf("Windows help should contain https, chocolatey, or choco reference: %s", help)
				}
			}
		})
	}
}

func containsAny(str string, substrings []string) bool {
	for _, sub := range substrings {
		if len(str) >= len(sub) {
			for i := 0; i <= len(str)-len(sub); i++ {
				if str[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// TestCheckerForClusterType verifies the type→requirements dispatch WITHOUT
// invoking CheckForClusterType: that function runs real installers in
// non-interactive mode, and an earlier version of this test did exactly that —
// it downloaded terraform onto CI runners and failed on `gcloud components
// install` (CI regression). Only the pure mapping is unit-testable.
func TestCheckerForClusterType(t *testing.T) {
	names := func(c *PrerequisiteChecker) []string {
		var out []string
		for _, r := range c.requirements {
			out = append(out, r.Name)
		}
		return out
	}

	cases := []struct {
		clusterType models.ClusterType
		want        []string
	}{
		{models.ClusterTypeK3d, []string{"Docker", "k3d", "helm"}},
		{models.ClusterType(""), []string{"Docker", "k3d", "helm"}},
		{models.ClusterTypeEKS, []string{"terraform", "AWS CLI"}},
		{models.ClusterTypeGKE, []string{"terraform", "gcloud", "gke-gcloud-auth-plugin"}},
	}
	for _, tc := range cases {
		checker := checkerForClusterType(tc.clusterType)
		if checker == nil {
			t.Fatalf("checkerForClusterType(%q) = nil, want a requirement set", tc.clusterType)
			continue
		}
		got := names(checker)
		if len(got) != len(tc.want) {
			t.Fatalf("checkerForClusterType(%q) = %v, want %v", tc.clusterType, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("checkerForClusterType(%q)[%d] = %s, want %s", tc.clusterType, i, got[i], tc.want[i])
			}
		}
		// Every requirement must be fully wired — a nil func would panic the
		// installer flow at runtime.
		for _, r := range checker.requirements {
			if r.IsInstalled == nil || r.Install == nil || r.InstallHelp == nil {
				t.Errorf("%s/%s: requirement funcs must all be set", tc.clusterType, r.Name)
			}
		}
	}

	if checkerForClusterType("unknown") != nil {
		t.Error("unknown types must return nil (gate passes, provider factory rejects)")
	}
}

func TestCheckAllWithMissingTools(t *testing.T) {
	checker := NewPrerequisiteChecker()

	// Requirements order: Docker(0), k3d(1), helm(2).
	checker.requirements[0].IsInstalled = func() bool { return false } // Docker - missing
	checker.requirements[1].IsInstalled = func() bool { return false } // k3d - missing
	checker.requirements[2].IsInstalled = func() bool { return true }  // helm - installed

	allPresent, missing := checker.CheckAll()

	if allPresent {
		t.Error("Expected allPresent to be false when tools are missing")
	}

	if len(missing) != 2 {
		t.Errorf("Expected 2 missing tools, got %d", len(missing))
	}

	expectedMissing := []string{"Docker", "k3d"}
	for i, tool := range missing {
		if tool != expectedMissing[i] {
			t.Errorf("Expected missing tool %d to be %s, got %s", i, expectedMissing[i], tool)
		}
	}
}

func TestCheckAllWithAllTools(t *testing.T) {
	checker := NewPrerequisiteChecker()

	for i := range checker.requirements {
		checker.requirements[i].IsInstalled = func() bool { return true }
	}

	allPresent, missing := checker.CheckAll()

	if !allPresent {
		t.Error("Expected allPresent to be true when all tools are present")
	}

	if len(missing) != 0 {
		t.Errorf("Expected no missing tools, got %d: %v", len(missing), missing)
	}
}
