package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Accessor inspects an existing cluster through a Kubernetes client. It never
// creates or mutates clusters — only reads health and capacity.
type Accessor struct {
	clientset kubernetes.Interface
}

// NewAccessorForConfig builds an Accessor from a rest.Config.
func NewAccessorForConfig(config *rest.Config) (*Accessor, error) {
	if config == nil {
		return nil, fmt.Errorf("rest.Config cannot be nil")
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &Accessor{clientset: cs}, nil
}

// Health is a snapshot of whether a cluster is reachable and ready.
type Health struct {
	Reachable     bool
	NodesTotal    int
	NodesReady    int
	ServerVersion string
}

// Ready reports whether the cluster is reachable and has at least one ready node.
func (h Health) Ready() bool { return h.Reachable && h.NodesReady > 0 }

// CheckHealth verifies the cluster is reachable and reports node readiness.
// A List error means the cluster is unreachable (Reachable=false).
func (a *Accessor) CheckHealth(ctx context.Context) (Health, error) {
	nodes, err := a.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return Health{Reachable: false}, fmt.Errorf("cluster is not reachable: %w", err)
	}

	h := Health{Reachable: true, NodesTotal: len(nodes.Items)}
	for i := range nodes.Items {
		if nodeReady(&nodes.Items[i]) {
			h.NodesReady++
		}
	}

	// Server version is best-effort; a failure here doesn't make the cluster unhealthy.
	if v, verr := a.clientset.Discovery().ServerVersion(); verr == nil && v != nil {
		h.ServerVersion = v.String()
	}
	return h, nil
}

// Requirements is the minimum allocatable capacity an install needs.
type Requirements struct {
	CPUMillis int64 // milli-CPU (e.g. 6000 = 6 cores)
	MemBytes  int64 // bytes
}

// Resources is the allocatable capacity summed across ready nodes.
type Resources struct {
	AllocatableCPUMillis int64
	AllocatableMemBytes  int64
}

// CheckResources sums allocatable CPU/memory across ready nodes and reports
// whether it meets req. `sufficient` is false when the cluster is too small.
func (a *Accessor) CheckResources(ctx context.Context, req Requirements) (res Resources, sufficient bool, err error) {
	nodes, err := a.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return Resources{}, false, fmt.Errorf("cluster is not reachable: %w", err)
	}

	for i := range nodes.Items {
		n := &nodes.Items[i]
		if !nodeReady(n) {
			continue // only count capacity we can actually schedule on
		}
		res.AllocatableCPUMillis += n.Status.Allocatable.Cpu().MilliValue()
		res.AllocatableMemBytes += n.Status.Allocatable.Memory().Value()
	}

	sufficient = res.AllocatableCPUMillis >= req.CPUMillis && res.AllocatableMemBytes >= req.MemBytes
	return res, sufficient, nil
}

// nodeReady reports whether a node's Ready condition is True.
func nodeReady(n *corev1.Node) bool {
	for _, cond := range n.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}
