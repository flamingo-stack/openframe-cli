package eks

import (
	"encoding/base64"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/providers/kubeconfigmerge"
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// EKS kubeconfig entries carry no static credentials: authentication runs
// through the client-go exec plugin (`aws eks get-token`), so tokens are
// short-lived and minted from the operator's AWS identity on every call.

// execArgs builds the aws exec-plugin argv for a cluster record.
func execArgs(rec tfengine.Record) []string {
	args := []string{"eks", "get-token", "--cluster-name", rec.Name, "--region", rec.Region, "--output", "json"}
	if rec.Profile != "" {
		args = append(args, "--profile", rec.Profile)
	}
	return args
}

func execConfig(rec tfengine.Record) *clientcmdapi.ExecConfig {
	return &clientcmdapi.ExecConfig{
		APIVersion:      "client.authentication.k8s.io/v1beta1",
		Command:         "aws",
		Args:            execArgs(rec),
		InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
	}
}

// caData decodes the base64 CA bundle the EKS module outputs.
func caData(rec tfengine.Record) ([]byte, error) {
	ca, err := base64.StdEncoding.DecodeString(rec.CACert)
	if err != nil {
		return nil, fmt.Errorf("decoding cluster CA for %s: %w", rec.Name, err)
	}
	return ca, nil
}

// kubeconfigFor renders an in-memory kubeconfig with a single context named
// after the cluster — the plain name (not an ARN) so the rest of the CLI can
// resolve it by exact match.
func kubeconfigFor(rec tfengine.Record) (*clientcmdapi.Config, error) {
	ca, err := caData(rec)
	if err != nil {
		return nil, err
	}
	cfg := clientcmdapi.NewConfig()
	cfg.Clusters[rec.Name] = &clientcmdapi.Cluster{
		Server:                   rec.Endpoint,
		CertificateAuthorityData: ca,
	}
	cfg.AuthInfos[rec.Name] = &clientcmdapi.AuthInfo{Exec: execConfig(rec)}
	cfg.Contexts[rec.Name] = &clientcmdapi.Context{Cluster: rec.Name, AuthInfo: rec.Name}
	cfg.CurrentContext = rec.Name
	return cfg, nil
}

// restConfigFor builds a rest.Config straight from the record — no kubeconfig
// file round-trip needed.
func restConfigFor(rec tfengine.Record) (*rest.Config, error) {
	ca, err := caData(rec)
	if err != nil {
		return nil, err
	}
	return &rest.Config{
		Host:            rec.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{CAData: ca},
		ExecProvider:    execConfig(rec),
	}, nil
}

// mergeIntoDefaultKubeconfig writes the cluster's context into the user's
// kubeconfig (honoring $KUBECONFIG) and switches the current context to it —
// the same post-create behavior the k3d provider gets from k3d itself.
// It refuses to overwrite a same-named context that points at a DIFFERENT
// server (see the GKE twin for rationale). The actual merge/no-clobber logic
// lives in the shared kubeconfigmerge package so EKS and GKE can't drift.
func mergeIntoDefaultKubeconfig(rec tfengine.Record) error {
	generated, err := kubeconfigFor(rec)
	if err != nil {
		return err
	}
	return kubeconfigmerge.MergeIntoDefault(kubeconfigmerge.Entry{
		Name:     rec.Name,
		Endpoint: rec.Endpoint,
		Cluster:  generated.Clusters[rec.Name],
		AuthInfo: generated.AuthInfos[rec.Name],
		Context:  generated.Contexts[rec.Name],
	})
}

// removeFromDefaultKubeconfig drops the cluster's context after a destroy —
// but ONLY when the entry still points at OUR endpoint (see the GKE twin).
// Best-effort: a missing entry is not an error. Delegates to the shared
// kubeconfigmerge package so the removal semantics can't diverge from GKE.
func removeFromDefaultKubeconfig(rec tfengine.Record) error {
	return kubeconfigmerge.RemoveFromDefault(rec.Name, rec.Endpoint)
}
