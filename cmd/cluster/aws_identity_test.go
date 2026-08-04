package cluster

import (
	"context"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
)

const stsIdentityJSON = `{"UserId":"AIDAEXAMPLE","Account":"123456789012","Arn":"arn:aws:iam::123456789012:user/dev"}`

// stubAWSIdentitySeams pins the interactivity probe and the confirmation
// prompt. It returns a pointer to the last prompt message ("" = never asked).
func stubAWSIdentitySeams(t *testing.T, interactive, answer bool, answerErr error) *string {
	t.Helper()
	asked := ""
	origConfirm, origInteractive := confirmAWSIdentityFn, awsInteractiveFn
	confirmAWSIdentityFn = func(message string, _ bool) (bool, error) {
		asked = message
		return answer, answerErr
	}
	awsInteractiveFn = func() bool { return interactive }
	t.Cleanup(func() { confirmAWSIdentityFn, awsInteractiveFn = origConfirm, origInteractive })
	return &asked
}

func TestConfirmAWSIdentity_NonInteractiveAnnouncesWithoutPrompt(t *testing.T) {
	asked := stubAWSIdentitySeams(t, false, false, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})

	if err := confirmAWSIdentity(context.Background(), mock, "staging"); err != nil {
		t.Fatalf("non-interactive with working credentials must proceed, got %v", err)
	}
	if *asked != "" {
		t.Fatalf("non-interactive mode must never prompt, asked: %q", *asked)
	}
}

func TestConfirmAWSIdentity_ProfileFlagReachesSTS(t *testing.T) {
	stubAWSIdentitySeams(t, false, false, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})

	if err := confirmAWSIdentity(context.Background(), mock, "staging"); err != nil {
		t.Fatal(err)
	}
	want := "aws sts get-caller-identity --output json --profile staging"
	if got := mock.GetLastCommand(); got != want {
		t.Fatalf("sts argv = %q, want %q", got, want)
	}

	// The default chain must NOT sprout a --profile flag.
	mock.Reset()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})
	if err := confirmAWSIdentity(context.Background(), mock, ""); err != nil {
		t.Fatal(err)
	}
	if got := mock.GetLastCommand(); got != "aws sts get-caller-identity --output json" {
		t.Fatalf("default-chain sts argv = %q, want no --profile", got)
	}
}

func TestConfirmAWSIdentity_InteractiveConfirmProceeds(t *testing.T) {
	asked := stubAWSIdentitySeams(t, true, true, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})

	if err := confirmAWSIdentity(context.Background(), mock, "staging"); err != nil {
		t.Fatalf("confirmed identity must proceed, got %v", err)
	}
	// The prompt must show WHO is about to be used: profile, account and ARN.
	for _, want := range []string{"profile 'staging'", "123456789012", "arn:aws:iam::123456789012:user/dev"} {
		if !strings.Contains(*asked, want) {
			t.Fatalf("prompt %q must contain %q", *asked, want)
		}
	}
}

func TestConfirmAWSIdentity_InteractiveDeclineAbortsWithAlternatives(t *testing.T) {
	stubAWSIdentitySeams(t, true, false, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})
	mock.SetResponse("configure list-profiles", &executor.CommandResult{Stdout: "dev\nstaging\n"})

	err := confirmAWSIdentity(context.Background(), mock, "staging")
	if err == nil {
		t.Fatal("a declined identity must abort the operation")
	}
	for _, want := range []string{"not confirmed", "--profile", "dev", "staging"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("decline error %q must contain %q", err, want)
		}
	}
}

func TestConfirmAWSIdentity_NoConfigurationPointsAtTheOfficialGuide(t *testing.T) {
	stubAWSIdentitySeams(t, true, true, nil)
	mock := executor.NewMockCommandExecutor()
	// Everything fails: no credentials, no profiles — nothing is configured.
	mock.SetShouldFail(true, "Unable to locate credentials")

	err := confirmAWSIdentity(context.Background(), mock, "")
	if err == nil {
		t.Fatal("expected an error when no AWS configuration exists")
	}
	for _, want := range []string{"no usable AWS configuration", awsConfigDocsURL, "--profile"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("no-config error %q must contain %q", err, want)
		}
	}
}

func TestConfirmAWSIdentity_BrokenProfileNamesTheWorkingOnes(t *testing.T) {
	stubAWSIdentitySeams(t, true, true, nil)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{ExitCode: 1, Stderr: "could not be found"})
	mock.SetResponse("configure list-profiles", &executor.CommandResult{Stdout: "dev\nprod\n"})

	err := confirmAWSIdentity(context.Background(), mock, "typo")
	if err == nil {
		t.Fatal("expected an error for an unusable profile")
	}
	for _, want := range []string{"profile 'typo'", "cannot authenticate", "dev", "prod", awsConfigDocsURL} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("broken-profile error %q must contain %q", err, want)
		}
	}
}

func TestConfirmAWSIdentity_PromptErrorSurfaces(t *testing.T) {
	stubAWSIdentitySeams(t, true, false, context.Canceled)
	mock := executor.NewMockCommandExecutor()
	mock.SetResponse("sts get-caller-identity", &executor.CommandResult{Stdout: stsIdentityJSON})

	if err := confirmAWSIdentity(context.Background(), mock, ""); err == nil {
		t.Fatal("a failed prompt read must surface, not silently proceed")
	}
}
