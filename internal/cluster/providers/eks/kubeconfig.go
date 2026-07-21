package eks

import (
	"encoding/base64"
	"fmt"

	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
// server (see the GKE twin for rationale).
func mergeIntoDefaultKubeconfig(rec tfengine.Record) error {
	pathOpts := clientcmd.NewDefaultPathOptions()
	existing, err := pathOpts.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("loading kubeconfig: %w", err)
	}
	if prior, ok := existing.Contexts[rec.Name]; ok {
		if cluster, ok := existing.Clusters[prior.Cluster]; ok && cluster.Server != rec.Endpoint {
			return fmt.Errorf("kubeconfig context '%s' already exists and points at %s — refusing to overwrite it; rename the existing context or pick another cluster name", rec.Name, cluster.Server)
		}
	}
	generated, err := kubeconfigFor(rec)
	if err != nil {
		return err
	}
	existing.Clusters[rec.Name] = generated.Clusters[rec.Name]
	existing.AuthInfos[rec.Name] = generated.AuthInfos[rec.Name]
	existing.Contexts[rec.Name] = generated.Contexts[rec.Name]
	existing.CurrentContext = rec.Name
	if err := clientcmd.ModifyConfig(pathOpts, *existing, true); err != nil {
		return fmt.Errorf("writing kubeconfig: %w", err)
	}
	return nil
}

// removeFromDefaultKubeconfig drops the cluster's context after a destroy.
// Best-effort: a missing entry is not an error.
func removeFromDefaultKubeconfig(name string) error {
	pathOpts := clientcmd.NewDefaultPathOptions()
	existing, err := pathOpts.GetStartingConfig()
	if err != nil {
		return err
	}
	delete(existing.Clusters, name)
	delete(existing.AuthInfos, name)
	delete(existing.Contexts, name)
	if existing.CurrentContext == name {
		existing.CurrentContext = ""
	}
	return clientcmd.ModifyConfig(pathOpts, *existing, true)
}
