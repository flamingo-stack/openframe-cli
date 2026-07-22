package terraform

import (
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEstimateMonthlyCost(t *testing.T) {
	planJSON := []byte(`{"format_version":"1.2"}`)

	t.Run("parses the infracost total", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("infracost breakdown", &executor.CommandResult{
			ExitCode: 0,
			Stdout:   `{"totalMonthlyCost":"142.53","currency":"USD","projects":[]}`,
		})

		cost, err := EstimateMonthlyCost(context.Background(), mock, planJSON)
		require.NoError(t, err)
		assert.Equal(t, "142.53 USD", cost)
		assert.True(t, mock.WasCommandExecuted("infracost breakdown"), "must invoke infracost breakdown")
	})

	t.Run("missing currency defaults to USD", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("infracost breakdown", &executor.CommandResult{
			ExitCode: 0, Stdout: `{"totalMonthlyCost":"99.00"}`,
		})
		cost, err := EstimateMonthlyCost(context.Background(), mock, planJSON)
		require.NoError(t, err)
		assert.Equal(t, "99.00 USD", cost)
	})

	t.Run("infracost failure is an error, never a made-up number", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "No INFRACOST_API_KEY environment variable is set")
		_, err := EstimateMonthlyCost(context.Background(), mock, planJSON)
		assert.Error(t, err)
	})

	t.Run("unparseable output is an error", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("infracost breakdown", &executor.CommandResult{ExitCode: 0, Stdout: "not json"})
		_, err := EstimateMonthlyCost(context.Background(), mock, planJSON)
		assert.Error(t, err)
	})

	t.Run("empty total is an error", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("infracost breakdown", &executor.CommandResult{ExitCode: 0, Stdout: `{"currency":"USD"}`})
		_, err := EstimateMonthlyCost(context.Background(), mock, planJSON)
		assert.Error(t, err)
	})

	t.Run("empty plan JSON is an error", func(t *testing.T) {
		_, err := EstimateMonthlyCost(context.Background(), executor.NewMockCommandExecutor(), nil)
		assert.Error(t, err)
	})
}
