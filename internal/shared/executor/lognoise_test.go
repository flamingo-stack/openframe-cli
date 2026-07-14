package executor

import (
	"errors"
	"strings"
	"testing"
)

// k3dFailureStderr is the real-world shape of a failed `k3d cluster create`:
// a wall of INFO progress records with the actual reason in FATA at the end.
const k3dFailureStderr = `INFO[0000] Using config file /tmp/k3d-config.yaml
INFO[0000] Prep: Network
INFO[0001] Created network 'k3d-openframe-dev'
INFO[0002] Created image volume k3d-openframe-dev-images
WARN[0003] No node filter specified
INFO[0004] Starting new tools node...
ERRO[0060] Failed Cluster Start: Failed to start server k3d-openframe-dev-server-0
FATA[0060] Cluster creation FAILED, all changes have been rolled back!`

func TestStripLogNoise_DropsInfoKeepsWarningsAndErrors(t *testing.T) {
	got := stripLogNoise(k3dFailureStderr)

	if strings.Contains(got, "INFO[") {
		t.Errorf("INFO records must be stripped, got:\n%s", got)
	}
	for _, want := range []string{"WARN[0003]", "ERRO[0060]", "FATA[0060]"} {
		if !strings.Contains(got, want) {
			t.Errorf("%s must survive the strip, got:\n%s", want, got)
		}
	}
}

func TestStripLogNoise_DropsLogfmtInfo(t *testing.T) {
	in := `time="2026-07-14T10:00:00Z" level=info msg="Prep: Network"
time="2026-07-14T10:00:01Z" level=error msg="port already allocated"`
	got := stripLogNoise(in)

	if strings.Contains(got, "level=info") {
		t.Errorf("logfmt info records must be stripped, got:\n%s", got)
	}
	if !strings.Contains(got, "port already allocated") {
		t.Errorf("logfmt error record must survive, got:\n%s", got)
	}
}

func TestStripLogNoise_PlainOutputUntouched(t *testing.T) {
	in := "Error: INSTALLATION FAILED: context deadline exceeded\nsee logs for details"
	if got := stripLogNoise(in); got != in {
		t.Errorf("non-logrus output must pass through untouched, got:\n%s", got)
	}
}

// TestErrorDetail_AllNoiseFallsBackToRaw: when the child only logged INFO
// before dying, the noise is the only detail there is — keep it.
func TestErrorDetail_AllNoiseFallsBackToRaw(t *testing.T) {
	in := "INFO[0000] Using config file\nINFO[0001] Prep: Network"
	if got := errorDetail(in); got != strings.TrimSpace(in) {
		t.Errorf("all-noise stderr must fall back to raw text, got:\n%s", got)
	}
}

// TestCommandError_FiltersK3dProgressWall locks the user-visible behavior: the
// error a failed k3d command produces names the failure, not 8 progress lines.
func TestCommandError_FiltersK3dProgressWall(t *testing.T) {
	err := &CommandError{
		Command:  "k3d cluster create --config /tmp/k3d-config.yaml",
		ExitCode: 1,
		Stderr:   k3dFailureStderr,
		cause:    errors.New("exit status 1"),
	}
	msg := err.Error()

	if strings.Contains(msg, "INFO[") {
		t.Errorf("INFO records must not reach the error message, got:\n%s", msg)
	}
	if !strings.Contains(msg, "FATA[0060] Cluster creation FAILED") {
		t.Errorf("the FATA line must reach the error message, got:\n%s", msg)
	}
	// The full unfiltered text stays available on the field for verbose paths.
	if !strings.Contains(err.Stderr, "INFO[0000]") {
		t.Error("Stderr field must keep the full unfiltered output")
	}
}

func TestWSLError_FiltersLogNoise(t *testing.T) {
	err := &WSLError{
		Operation: "executing k3d",
		ExitCode:  1,
		Stderr:    k3dFailureStderr,
	}
	msg := err.Error()

	if strings.Contains(msg, "INFO[") {
		t.Errorf("INFO records must not reach the WSL error message, got:\n%s", msg)
	}
	if !strings.Contains(msg, "FATA[0060]") {
		t.Errorf("the FATA line must reach the WSL error message, got:\n%s", msg)
	}
}
