package cluster

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

func TestValidateGKEProjectFlag(t *testing.T) {
	list := &executor.CommandResult{ExitCode: 0, Stdout: "tenant-y0\nshared-abc\nprod-42\n"}

	t.Run("project in the list passes", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud projects list", list)
		if err := validateGKEProjectFlag(context.Background(), mock, "shared-abc"); err != nil {
			t.Fatalf("expected nil for a listed project, got %v", err)
		}
	})

	t.Run("project not in the list errors with the valid options", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetResponse("gcloud projects list", list)
		err := validateGKEProjectFlag(context.Background(), mock, "typo-proj")
		if err == nil {
			t.Fatal("expected an error for an unlisted project")
		}
		if !strings.Contains(err.Error(), "typo-proj") || !strings.Contains(err.Error(), "tenant-y0") {
			t.Fatalf("error should name the bad project and list valid options, got: %v", err)
		}
	})

	t.Run("best-effort: a gcloud listing failure does not block create", func(t *testing.T) {
		mock := executor.NewMockCommandExecutor()
		mock.SetShouldFail(true, "network down")
		if err := validateGKEProjectFlag(context.Background(), mock, "anything"); err != nil {
			t.Fatalf("a gcloud failure must defer to the provider preflight, not block; got %v", err)
		}
	})
}
