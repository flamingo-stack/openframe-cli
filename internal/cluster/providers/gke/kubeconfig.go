package gke

import (
	tfengine "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/terraform"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	kubeconfighelper "github.com/flamingo-stack/openframe-cli/internal/cluster/providers/kubeconfig"
)

// GKE kubeconfig entries carry no static credentials: authentication runs
// through the client-go exec plugin (gke-gcloud-auth-plugin), so tokens are
// short-lived and minted from the operator's gcloud identity on every call.

func execConfig() *clientcmdapi.ExecConfig {
	return &clientcmdapi.ExecConfig{
		APIVersion:         "client.authentication.k8s.io/v1beta1",
		Command:            "gke-gcloud-auth-plugin",
		InteractiveMode:    clientcmdapi.NeverExecInteractiveMode,
		ProvideClusterInfo: true,
	}
}

// caData decodes the base64 CA bundle the GKE module outputs.
func caData(rec tfengine.Record) ([]byte, error) {
	return kubeconfighelper.CAData(rec)
}

// kubeconfigFor renders an in-memory kubeconfig with a single context named
// after the cluster — the plain name so the rest of the CLI resolves it by
// exact match.
func kubeconfigFor(rec tfengine.Record) (*clientcmdapi.Config, error) {
	return kubeconfighelper.KubeconfigFor(rec, func(tfengine.Record) *clientcmdapi.ExecConfig {
		return execConfig()
	})
}

// restConfigFor builds a rest.Config straight from the record.
func restConfigFor(rec tfengine.Record) (*rest.Config, error) {
	return kubeconfighelper.RestConfigFor(rec, func(tfengine.Record) *clientcmdapi.ExecConfig {
		return execConfig()
	})
}

// mergeIntoDefaultKubeconfig writes the cluster's context into the user's
// kubeconfig (honoring $KUBECONFIG) and switches the current context to it.
// It refuses to overwrite a same-named context that points at a DIFFERENT
// server: that context belongs to something else (another cluster, another
// tool) and silently clobbering it would break the user's access to it.
func mergeIntoDefaultKubeconfig(rec tfengine.Record) error {
	return kubeconfighelper.MergeIntoDefaultKubeconfig(rec, func(tfengine.Record) *clientcmdapi.ExecConfig {
		return execConfig()
	})
}

// removeFromDefaultKubeconfig drops the cluster's context after a destroy —
// but ONLY when the entry still points at OUR endpoint. If the user repointed
// or recreated a same-named context toward another server since the create,
// it is no longer ours to delete (the create-side no-clobber guard's mirror).
func removeFromDefaultKubeconfig(rec tfengine.Record) error {
	return kubeconfighelper.RemoveFromDefaultKubeconfig(rec)
}
