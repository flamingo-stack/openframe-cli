package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHelmInstaller_And_InstallHelp(t *testing.T) {
	h := NewHelmInstaller()
	require.NotNil(t, h)

	help := h.GetInstallHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "helm")
	assert.Contains(t, help, "https://helm.sh", "install help must point to the official docs")
}

func TestCommandExists(t *testing.T) {
	// The Go toolchain is running these tests, so `go` is guaranteed on PATH.
	assert.True(t, commandExists("go"))
	assert.False(t, commandExists("definitely-not-a-real-command-9z8x7"))
}

func TestContainsPath(t *testing.T) {
	cases := []struct {
		name    string
		pathEnv string
		dir     string
		want    bool
	}{
		{"present in the middle", "/a;/b;/c", "/b", true},
		{"present at the end", "/a;/b", "/b", true},
		{"absent", "/a;/b", "/z", false},
		{"empty path", "", "/b", false},
		{"case-insensitive match", `C:\Users\bin`, `c:\users\bin`, true},
		{"trims surrounding spaces", "/a ; /b ; /c", "/b", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, containsPath(tc.pathEnv, tc.dir))
		})
	}
}
