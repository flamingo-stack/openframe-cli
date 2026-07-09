package argocd

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func pod(phase corev1.PodPhase, ready bool) *corev1.Pod {
	cond := corev1.ConditionFalse
	if ready {
		cond = corev1.ConditionTrue
	}
	return &corev1.Pod{Status: corev1.PodStatus{
		Phase:      phase,
		Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: cond}},
	}}
}

func TestIsPodReady(t *testing.T) {
	cases := []struct {
		name string
		pod  *corev1.Pod
		want bool
	}{
		{"succeeded job pod is ready", pod(corev1.PodSucceeded, false), true},
		{"running + Ready=True", pod(corev1.PodRunning, true), true},
		{"running + Ready=False", pod(corev1.PodRunning, false), false},
		{"pending", pod(corev1.PodPending, false), false},
		{"failed", pod(corev1.PodFailed, false), false},
		{"running, no conditions", &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isPodReady(c.pod); got != c.want {
				t.Fatalf("isPodReady = %v, want %v", got, c.want)
			}
		})
	}
}
