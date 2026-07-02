package app

import (
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/chart/providers/argocd"
	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	sharedErrors "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// getAccessCmd returns the access subcommand.
func getAccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Print ArgoCD admin credentials and how to open the UI",
		Long: `Show how to sign in to the OpenFrame control plane (ArgoCD).

Reads the initial admin password from the cluster and prints the username,
password, and the command to open the ArgoCD UI locally.

Examples:
  openframe app access
  openframe app access --context k3d-openframe-dev`,
		RunE:        runAccessCommand,
		Annotations: map[string]string{"readonly": "true"},
	}
	cmd.Flags().String("context", "", "Kube-context to use (defaults to the current context)")
	addOutputFlag(cmd)
	return cmd
}

func runAccessCommand(cmd *cobra.Command, _ []string) error {
	verbose := getVerboseFlag(cmd)
	contextName, _ := cmd.Flags().GetString("context")
	format, err := outputFormat(cmd)
	if err != nil {
		return sharedErrors.HandleGlobalError(err, verbose)
	}

	mgr, err := newArgoCDManager(contextName, verbose)
	if err != nil {
		return sharedErrors.HandleGlobalError(fmt.Errorf("could not connect to the cluster: %w", err), verbose)
	}

	password, err := mgr.AdminPassword(cmd.Context())
	if err != nil {
		return sharedErrors.HandleGlobalError(
			fmt.Errorf("could not read the ArgoCD admin password — is OpenFrame installed? (%w)", err), verbose)
	}

	if format == "json" {
		return printJSON(struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{Username: "admin", Password: password})
	}
	printAccess(password)
	return nil
}

// resolveRestConfig builds a rest.Config for the given kube-context (empty means
// the current context). Shared by the status and access commands.
func resolveRestConfig(contextName string) (*rest.Config, error) {
	return k8s.RestConfigForContext(k8s.DefaultKubeconfigPath(), contextName)
}

// newArgoCDManager builds an ArgoCD manager bound to the given context.
func newArgoCDManager(contextName string, verbose bool) (*argocd.Manager, error) {
	cfg, err := resolveRestConfig(contextName)
	if err != nil {
		return nil, err
	}
	return argocd.NewManagerWithConfig(executor.NewRealCommandExecutor(false, verbose), cfg)
}

// printAccess renders the ArgoCD sign-in details and how to reach the UI.
func printAccess(password string) {
	pterm.DefaultSection.Println("ArgoCD access")
	pterm.Printf("  Username: admin\n")
	pterm.Printf("  Password: %s\n", password)
	pterm.Info.Println("Open the ArgoCD UI:")
	pterm.Printf("  1. kubectl port-forward -n argocd svc/argocd-server 8080:443\n")
	pterm.Printf("  2. open https://localhost:8080\n")
}
