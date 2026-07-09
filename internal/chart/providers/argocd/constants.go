package argocd

// ArgoCD install identity — single source of truth for the pinned chart and the
// names the CLI installs it under. A chart bump or rename is a one-line change
// here instead of a scattered find-and-replace.
//
// NOTE: ArgoCDNamespace is coupled with `fullnameOverride: argocd` in
// argocd-values.yaml; keep the two in sync.
const (
	ArgoCDNamespace    = "argocd"
	ArgoCDReleaseName  = "argo-cd"
	ArgoCDChartRef     = "argo/argo-cd"
	ArgoCDChartVersion = "10.1.1"
	ArgoHelmRepoURL    = "https://argoproj.github.io/argo-helm"
)

// ArgoCD Application health and sync status values. These are compared for
// equality to decide when the install is complete, so a typo would silently
// break the completion logic — always use these constants, never the raw string.
const (
	ArgoCDHealthHealthy     = "Healthy"
	ArgoCDHealthProgressing = "Progressing"
	ArgoCDHealthDegraded    = "Degraded"
	ArgoCDHealthMissing     = "Missing"
	ArgoCDSyncSynced        = "Synced"
	ArgoCDSyncOutOfSync     = "OutOfSync"
	// ArgoCDStatusUnknown is reported for both health and sync when ArgoCD hasn't
	// determined a status yet.
	ArgoCDStatusUnknown = "Unknown"
)
