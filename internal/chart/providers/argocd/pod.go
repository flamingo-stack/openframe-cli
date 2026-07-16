package argocd

import corev1 "k8s.io/api/core/v1"

// isPodReady checks if a pod has the Ready condition set to True
// Completed Job pods (like argocd-redis-secret-init) are considered "ready" since they finished successfully
func isPodReady(pod *corev1.Pod) bool {
	// Completed pods (from Jobs) are considered ready - they finished their work successfully
	if pod.Status.Phase == corev1.PodSucceeded {
		return true
	}

	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
