package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/discovery"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	sharedUI "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
	"github.com/pterm/pterm"
)

// awsConfigDocsURL is the official AWS guide for setting up the config and
// credential files — surfaced whenever no usable AWS configuration is found.
const awsConfigDocsURL = "https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html"

// Seams for hermetic tests: the confirmation reads the terminal and the
// interactivity probe inspects the real stdin/CI environment.
var (
	confirmAWSIdentityFn = sharedUI.ConfirmActionInteractive
	awsInteractiveFn     = func() bool { return !sharedUI.IsNonInteractive() }
)

// awsIdentity is the caller identity an EKS operation would run under, as
// reported by `aws sts get-caller-identity`.
type awsIdentity struct {
	Account string `json:"Account"`
	Arn     string `json:"Arn"`
}

// describeAWSSelection names the credential source for user-facing messages.
func describeAWSSelection(profile string) string {
	if profile == "" {
		return "default credentials (no profile)"
	}
	return fmt.Sprintf("profile '%s'", profile)
}

// confirmAWSIdentity resolves the AWS identity an EKS operation is about to
// use (--profile, or the default credential chain) and vets it with the user.
// Interactively the resolved account/ARN must be explicitly confirmed — the
// wrong profile provisions billed resources in the wrong account. In a
// non-interactive session it never prompts: an explicit --profile (or a
// working default chain) is the consent, but the account in use is announced
// so CI logs show whose account was billed. When nothing can authenticate,
// the error says what is missing and points at the official AWS
// configuration guide. The GKE twin of this flow is discovery.AuthFlow.
func confirmAWSIdentity(ctx context.Context, exec executor.CommandExecutor, profile string) error {
	args := []string{"sts", "get-caller-identity", "--output", "json"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	res, err := exec.Execute(ctx, "aws", args...)
	if err != nil {
		// Distinguish "nothing is configured at all" from "this selection is
		// broken": the first gets the setup guide, the second the working
		// alternatives. Listing profiles is best-effort — on error it just
		// means no alternatives to offer.
		profiles, _ := discovery.NewEKSDiscoverer(exec).Profiles(ctx)
		if profile == "" && len(profiles) == 0 {
			return fmt.Errorf("no usable AWS configuration found: the default credential chain cannot authenticate and no named profiles exist\n"+
				"Set up AWS credentials first — official guide: %s\n"+
				"Then re-run (optionally with --profile <name>)", awsConfigDocsURL)
		}
		hint := ""
		if len(profiles) > 0 {
			hint = fmt.Sprintf("\nAvailable profiles: %s", strings.Join(profiles, ", "))
		}
		return fmt.Errorf("AWS %s cannot authenticate (aws sts get-caller-identity failed): %w%s\n"+
			"To (re)configure credentials, see the official guide: %s",
			describeAWSSelection(profile), err, hint, awsConfigDocsURL)
	}

	// The identity is display-only, so a parse failure must not block the
	// operation — the authentication itself already succeeded above.
	var id awsIdentity
	if res != nil {
		_ = json.Unmarshal([]byte(res.Stdout), &id)
	}
	identity := "identity details unavailable"
	if id.Account != "" || id.Arn != "" {
		identity = fmt.Sprintf("account %s, %s", id.Account, id.Arn)
	}
	who := describeAWSSelection(profile)

	if !awsInteractiveFn() {
		pterm.Info.Printf("Using AWS %s — %s\n", who, identity)
		return nil
	}

	ok, err := confirmAWSIdentityFn(fmt.Sprintf("Use AWS %s — %s for this EKS operation?", who, identity), true)
	if err != nil {
		return fmt.Errorf("failed to read the AWS identity confirmation: %w", err)
	}
	if !ok {
		profiles, _ := discovery.NewEKSDiscoverer(exec).Profiles(ctx)
		hint := "re-run with --profile <name> to pick another identity"
		if len(profiles) > 0 {
			hint = fmt.Sprintf("%s (available: %s)", hint, strings.Join(profiles, ", "))
		}
		return fmt.Errorf("EKS operation aborted: AWS identity not confirmed — %s", hint)
	}
	return nil
}
