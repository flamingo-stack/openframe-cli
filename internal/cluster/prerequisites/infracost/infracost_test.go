package infracost

import (
	"path/filepath"
	"testing"
)

// The release archive's member layout is irregular: unix builds embed os/arch
// in the member name, windows ships a bare infracost.exe — and the install
// destination must match the OS's executable naming.
func TestArchiveMemberAndDest(t *testing.T) {
	binDir := filepath.Join("home", ".openframe", "bin")
	cases := []struct {
		goos, goarch string
		wantMember   string
		wantDest     string
	}{
		{"darwin", "arm64", "infracost-darwin-arm64", filepath.Join(binDir, "infracost")},
		{"linux", "amd64", "infracost-linux-amd64", filepath.Join(binDir, "infracost")},
		{"windows", "amd64", "infracost.exe", filepath.Join(binDir, "infracost.exe")},
	}
	for _, tc := range cases {
		t.Run(tc.goos+"/"+tc.goarch, func(t *testing.T) {
			member, dest := archiveMemberAndDest(tc.goos, tc.goarch, binDir)
			if member != tc.wantMember {
				t.Errorf("member = %q, want %q", member, tc.wantMember)
			}
			if dest != tc.wantDest {
				t.Errorf("dest = %q, want %q", dest, tc.wantDest)
			}
		})
	}
}

func TestGetInstallHelp_NonEmpty(t *testing.T) {
	if NewInstaller().GetInstallHelp() == "" {
		t.Fatal("infracost install help must not be empty")
	}
}
