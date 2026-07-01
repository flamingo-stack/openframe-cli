// Package target resolves and validates the cluster an `app install` should
// target: it runs the interactive context-selection flow, builds a client for
// the chosen context, and verifies the cluster is reachable, ready, and large
// enough before anything is installed (req 16).
//
// Its external seams (context loading, config building, cluster checking) are
// injectable so the whole flow is unit-testable without a real cluster.
package target

import (
	"context"
	"fmt"

	"github.com/flamingo-stack/openframe-cli/internal/k8s"
	"k8s.io/client-go/rest"
)

// clusterChecker is the subset of *k8s.Accessor used here (injected in tests).
type clusterChecker interface {
	CheckHealth(ctx context.Context) (k8s.Health, error)
	CheckResources(ctx context.Context, req k8s.Requirements) (k8s.Resources, bool, error)
}

// Selector resolves and validates an install target.
type Selector struct {
	Prompter       k8s.Prompter
	Requirements   k8s.Requirements
	KubeconfigPath string

	// Injectable seams; NewSelector wires the production implementations.
	loadContexts func(path string) ([]k8s.ContextInfo, string, error)
	buildConfig  func(path, ctxName string) (*rest.Config, error)
	newChecker   func(*rest.Config) (clusterChecker, error)
}

// NewSelector returns a Selector wired to the real kubeconfig/cluster.
func NewSelector(prompter k8s.Prompter, req k8s.Requirements) *Selector {
	return &Selector{
		Prompter:       prompter,
		Requirements:   req,
		KubeconfigPath: k8s.DefaultKubeconfigPath(),
		loadContexts:   k8s.LoadContexts,
		buildConfig:    k8s.RestConfigForContext,
		newChecker: func(c *rest.Config) (clusterChecker, error) {
			return k8s.NewAccessorForConfig(c)
		},
	}
}

// Result is a validated install target.
type Result struct {
	Context             string
	Config              *rest.Config
	Health              k8s.Health
	Resources           k8s.Resources
	ResourcesSufficient bool // false → the cluster is smaller than recommended (advisory)
}

// Select runs the full flow: pick a context, build its client, and confirm the
// cluster is reachable, ready, and has enough capacity. Any failure returns a
// plain-language error suitable for a non-technical user.
func (s *Selector) Select(ctx context.Context) (Result, error) {
	contexts, current, err := s.loadContexts(s.KubeconfigPath)
	if err != nil {
		return Result{}, fmt.Errorf("could not read kubeconfig %s: %w", s.KubeconfigPath, err)
	}

	chosen, err := k8s.SelectContext(contexts, current, s.Prompter)
	if err != nil {
		return Result{}, err
	}

	cfg, err := s.buildConfig(s.KubeconfigPath, chosen)
	if err != nil {
		return Result{}, err
	}

	checker, err := s.newChecker(cfg)
	if err != nil {
		return Result{}, err
	}

	health, err := checker.CheckHealth(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("cluster %q is not reachable — is it running? (%w)", chosen, err)
	}
	if !health.Ready() {
		return Result{}, fmt.Errorf("cluster %q has no ready nodes yet — wait for it to come up and try again", chosen)
	}

	// Resource sufficiency is advisory: report it so the caller can warn the
	// user, but don't block the install on a borderline cluster.
	res, sufficient, err := checker.CheckResources(ctx, s.Requirements)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Context:             chosen,
		Config:              cfg,
		Health:              health,
		Resources:           res,
		ResourcesSufficient: sufficient,
	}, nil
}
