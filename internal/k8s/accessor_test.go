package k8s

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func node(name string, ready bool, cpu, mem string) *corev1.Node {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(cpu),
				corev1.ResourceMemory: resource.MustParse(mem),
			},
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: status}},
		},
	}
}

func TestCheckHealth_CountsReadyNodes(t *testing.T) {
	cs := fake.NewSimpleClientset(
		node("ready-1", true, "4", "8Gi"),
		node("notready", false, "4", "8Gi"),
	)
	h, err := (&Accessor{clientset: cs}).CheckHealth(context.Background())
	require.NoError(t, err)
	assert.True(t, h.Reachable)
	assert.Equal(t, 2, h.NodesTotal)
	assert.Equal(t, 1, h.NodesReady)
	assert.True(t, h.Ready(), "one ready node means the cluster is ready")
}

func TestCheckHealth_Unreachable(t *testing.T) {
	cs := fake.NewSimpleClientset()
	cs.PrependReactor("list", "nodes", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("connection refused")
	})
	h, err := (&Accessor{clientset: cs}).CheckHealth(context.Background())
	require.Error(t, err)
	assert.False(t, h.Reachable)
	assert.False(t, h.Ready())
}

func TestCheckResources_SumsReadyNodesOnly(t *testing.T) {
	cs := fake.NewSimpleClientset(
		node("ready-1", true, "4", "8Gi"),
		node("ready-2", true, "2", "4Gi"),
		node("notready", false, "8", "16Gi"), // must be excluded
	)
	a := &Accessor{clientset: cs}

	// ready capacity = 6 CPU (6000m), 12Gi
	res, ok, err := a.CheckResources(context.Background(), Requirements{CPUMillis: 6000, MemBytes: 12 * 1024 * 1024 * 1024})
	require.NoError(t, err)
	assert.Equal(t, int64(6000), res.AllocatableCPUMillis)
	assert.Equal(t, int64(12*1024*1024*1024), res.AllocatableMemBytes)
	assert.True(t, ok, "exactly meets the requirement")

	// requiring more than available → insufficient
	_, ok, err = a.CheckResources(context.Background(), Requirements{CPUMillis: 8000, MemBytes: 1})
	require.NoError(t, err)
	assert.False(t, ok, "8 cores required but only 6 available")
}
