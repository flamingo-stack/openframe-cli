package eks

import (
	"fmt"
	"regexp"
	"strings"
)

// awsResourceArnRE pulls an AWS ARN out of a terraform error so an orphan can
// be named concretely.
var awsResourceArnRE = regexp.MustCompile(`arn:aws[a-z-]*:[^'"\s,)]+`)

// orphanFromInterruptedCreate detects the specific failure where terraform
// tries to create a resource that already exists in AWS (ResourceInUseException
// / *AlreadyExists*). This is the signature of a create interrupted (SIGINT)
// after the cloud API created a resource but before terraform saved it to
// state: the resource is real but state-invisible, so every resume collides
// with it and 'cluster delete' — which only knows state-tracked resources —
// cannot remove it. Returns human-readable remediation and true when this is
// that case, so the caller can replace the generic "re-run to resume" hint
// (which would loop forever here) with something actionable. (The GKE twin
// detects the HTTP 409 / alreadyExists equivalent.)
func orphanFromInterruptedCreate(err error, terraformDir string) (string, bool) {
	msg := err.Error()
	low := strings.ToLower(msg)
	if !strings.Contains(low, "resourceinuseexception") && !strings.Contains(low, "alreadyexists") {
		return "", false
	}
	resource := "the resource named in the error above"
	if m := awsResourceArnRE.FindString(msg); m != "" {
		resource = m
	}
	return fmt.Sprintf(
		"a resource already exists in AWS that terraform is not tracking:\n"+
			"  %s\n"+
			"This is the signature of a create that was interrupted after the resource was\n"+
			"created but before its state was saved — so resume keeps colliding with it and\n"+
			"'cluster delete' cannot remove it (delete only knows state-tracked resources).\n"+
			"Resolve it one of two ways, then re-run create:\n"+
			"  • delete the orphan in AWS (Console or the matching 'aws ... delete' command), or\n"+
			"  • import it into this cluster's state, e.g.:\n"+
			"      terraform -chdir=%s import <resource.address> <resource-id>",
		resource, terraformDir), true
}
