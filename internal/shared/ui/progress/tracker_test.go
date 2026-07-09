package progress

import (
	"errors"
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain silences pterm so tests don't render spinners/output. These tests
// exercise only the pure state machine (they never call Start/UpdateProgress,
// which spin up pterm spinners/bars).
func TestMain(m *testing.M) {
	pterm.DisableOutput()
	os.Exit(m.Run())
}

func steps() []Step {
	return []Step{
		{Name: "a", Weight: 1},
		{Name: "b", Weight: 1},
		{Name: "c", Weight: 1},
	}
}

func TestStepStatus_String(t *testing.T) {
	assert.Equal(t, "Pending", StepPending.String())
	assert.Equal(t, "Running", StepRunning.String())
	assert.Equal(t, "Completed", StepCompleted.String())
	assert.Equal(t, "Failed", StepFailed.String())
	assert.Equal(t, "Skipped", StepSkipped.String())
	assert.Equal(t, "Unknown", StepStatus(99).String())
}

func TestNewTracker_InitialState(t *testing.T) {
	tr := NewTracker("deploy", steps())
	assert.Equal(t, -1, tr.currentStep)
	for i, s := range tr.steps {
		assert.Equalf(t, StepPending, s.Status, "step %d should start pending", i)
	}
}

func TestStartStep(t *testing.T) {
	tr := NewTracker("deploy", steps())

	require.NoError(t, tr.StartStep(0))
	assert.Equal(t, 0, tr.currentStep)
	assert.Equal(t, StepRunning, tr.steps[0].Status)

	assert.Error(t, tr.StartStep(-1))
	assert.Error(t, tr.StartStep(99))
}

func TestStartStep_TransitionsPreviousToCompleted(t *testing.T) {
	tr := NewTracker("deploy", steps())
	require.NoError(t, tr.StartStep(0))
	require.NoError(t, tr.StartStep(1))
	assert.Equal(t, StepCompleted, tr.steps[0].Status, "starting the next step completes the running one")
	assert.Equal(t, StepRunning, tr.steps[1].Status)
}

func TestCompleteStep(t *testing.T) {
	tr := NewTracker("deploy", steps())
	require.NoError(t, tr.StartStep(0))
	require.NoError(t, tr.CompleteStep(0))
	assert.Equal(t, StepCompleted, tr.steps[0].Status)
	assert.Error(t, tr.CompleteStep(99))
}

func TestFailStep(t *testing.T) {
	tr := NewTracker("deploy", steps())
	boom := errors.New("boom")
	require.NoError(t, tr.FailStep(1, boom))
	assert.Equal(t, StepFailed, tr.steps[1].Status)
	assert.Equal(t, boom, tr.steps[1].Error)
	assert.Error(t, tr.FailStep(99, boom))
}

func TestSkipStep(t *testing.T) {
	tr := NewTracker("deploy", steps())
	require.NoError(t, tr.SkipStep(2, "not needed"))
	assert.Equal(t, StepSkipped, tr.steps[2].Status)
	assert.Error(t, tr.SkipStep(99, "x"))
}

func TestGetProgress_Weighted(t *testing.T) {
	tr := NewTracker("deploy", steps())
	assert.InDelta(t, 0.0, tr.GetProgress(), 0.001)

	require.NoError(t, tr.CompleteStep(0))
	assert.InDelta(t, 33.333, tr.GetProgress(), 0.01)

	require.NoError(t, tr.CompleteStep(1))
	require.NoError(t, tr.CompleteStep(2))
	assert.InDelta(t, 100.0, tr.GetProgress(), 0.001)
}

func TestGetProgress_ZeroWeightsIsZero(t *testing.T) {
	tr := NewTracker("deploy", []Step{{Name: "a"}, {Name: "b"}}) // Weight 0
	assert.Equal(t, 0.0, tr.GetProgress())
}
