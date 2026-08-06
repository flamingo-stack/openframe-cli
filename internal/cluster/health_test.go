package cluster

import (
	"context"
	"errors"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// The creation box used to print a hardcoded green "Ready" on the
// provisioner's exit code alone — identically over NotReady nodes or a
// cluster with no default StorageClass. These tests pin the verified
// rendering.

func node(name string, ready bool) *corev1.Node {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: status},
		}},
	}
}

func storageClass(name string, isDefault bool) *storagev1.StorageClass {
	sc := &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if isDefault {
		sc.Annotations = map[string]string{"storageclass.kubernetes.io/is-default-class": "true"}
	}
	return sc
}

func TestObserveClusterHealth(t *testing.T) {
	t.Run("all ready with default storage class", func(t *testing.T) {
		client := fake.NewSimpleClientset(node("a", true), node("b", true), storageClass("gp3", true))
		h := observeClusterHealth(context.Background(), client)
		assert.True(t, h.healthy())
		assert.Equal(t, 2, h.readyNodes)
		assert.Equal(t, 2, h.totalNodes)
	})

	t.Run("not-ready node breaks health", func(t *testing.T) {
		client := fake.NewSimpleClientset(node("a", true), node("b", false), storageClass("gp3", true))
		h := observeClusterHealth(context.Background(), client)
		assert.False(t, h.healthy())
		assert.Equal(t, 1, h.readyNodes)
		assert.Equal(t, 2, h.totalNodes)
	})

	t.Run("missing default storage class breaks health", func(t *testing.T) {
		// A StorageClass exists (legacy gp2) but none is default, so every
		// PVC would stay Pending — the box must not say Ready.
		client := fake.NewSimpleClientset(node("a", true), storageClass("gp2", false))
		h := observeClusterHealth(context.Background(), client)
		assert.False(t, h.healthy())
		assert.False(t, h.hasDefaultStorageClass)
	})

	t.Run("no nodes is not healthy", func(t *testing.T) {
		client := fake.NewSimpleClientset(storageClass("gp3", true))
		assert.False(t, observeClusterHealth(context.Background(), client).healthy())
	})
}

func TestSummaryStatusLines(t *testing.T) {
	t.Run("verified healthy renders Ready", func(t *testing.T) {
		status, nodes := summaryStatusLines(clusterHealth{readyNodes: 3, totalNodes: 3, hasDefaultStorageClass: true}, 3)
		assert.Contains(t, pterm.RemoveColorFromString(status), "Ready")
		assert.Equal(t, "3/3 Ready", nodes)
	})

	t.Run("unhealthy renders the observed fraction, never Ready", func(t *testing.T) {
		status, nodes := summaryStatusLines(clusterHealth{readyNodes: 0, totalNodes: 3, hasDefaultStorageClass: true}, 3)
		plain := pterm.RemoveColorFromString(status)
		assert.Contains(t, plain, "0/3 nodes Ready")
		assert.NotEqual(t, "Ready", plain)
		assert.Equal(t, "0/3 Ready", nodes)
	})

	t.Run("unreachable API says so instead of guessing", func(t *testing.T) {
		status, nodes := summaryStatusLines(clusterHealth{verifyErr: errors.New("dial tcp: timeout")}, 3)
		assert.Contains(t, pterm.RemoveColorFromString(status), "not verified")
		assert.Equal(t, "3 (configured)", nodes)
	})
}
