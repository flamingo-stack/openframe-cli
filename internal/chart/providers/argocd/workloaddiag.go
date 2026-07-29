package argocd

import (
	"bufio"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Diagnostics for a Degraded application's workloads: when an ArgoCD app applied
// all its manifests (Synced) but a pod is unhealthy (Degraded), this pulls the
// failing pod's crash logs and recent warning events so a hung install says WHY
// — not just which app. This is what turns a silent 40-minute timeout into an
// instant root cause (e.g. a CrashLooping pod's placeholder error).

const (
	diagLogTailLines = 15
	diagMaxEvents    = 5
	diagMaxLogChars  = 1200
	// crashLoopRestartThreshold: a container with at least this many restarts is
	// treated as a genuine CrashLoop (won't recover on its own), even if its
	// current waiting reason is momentarily something else.
	crashLoopRestartThreshold = 3
)

// terminalWaitingReasons are container "waiting" states that do not recover on
// their own — the workload is definitively stuck, so waiting for it is futile.
var terminalWaitingReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"ImagePullBackOff":           true,
	"ErrImagePull":               true,
	"InvalidImageName":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
}

// isImagePullReason reports whether a waiting reason is an image-pull failure,
// which usually means a private image without registry credentials.
func isImagePullReason(reason string) bool {
	return reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "InvalidImageName"
}

// containerIssue is one failing container found in a pod.
type containerIssue struct {
	pod       string
	container string
	image     string
	reason    string
	restarts  int32
	terminal  bool // will not recover on its own
}

// failingContainers returns the containers in a pod that are in an error state:
// waiting with an error reason (excluding the benign startup reasons), or
// terminated with a non-zero exit. Pure and testable.
func failingContainers(p corev1.Pod) []containerIssue {
	var issues []containerIssue
	statuses := append(append([]corev1.ContainerStatus{}, p.Status.InitContainerStatuses...), p.Status.ContainerStatuses...)
	for _, cs := range statuses {
		var reason string
		switch {
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" &&
			cs.State.Waiting.Reason != "PodInitializing" && cs.State.Waiting.Reason != "ContainerCreating":
			reason = cs.State.Waiting.Reason
		case cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0:
			reason = cs.State.Terminated.Reason
			if reason == "" {
				reason = fmt.Sprintf("Exited(%d)", cs.State.Terminated.ExitCode)
			}
		default:
			continue
		}
		issues = append(issues, containerIssue{
			pod:       p.Name,
			container: cs.Name,
			image:     cs.Image,
			reason:    reason,
			restarts:  cs.RestartCount,
			terminal:  terminalWaitingReasons[reason] || cs.RestartCount >= crashLoopRestartThreshold,
		})
	}
	return issues
}

// diagnoseFailingApps inspects the workloads of the given apps and returns a
// human diagnostic — the failing pod/container, why (waiting reason or crash), a
// registry-auth hint for image-pull failures, the crashed container's last log
// lines, and recent warning events. terminal is true when any workload is in a
// won't-recover state (CrashLoop / ImagePull / repeated crashes), letting the
// caller fail fast instead of waiting out the timeout. Best-effort throughout.
func (m *Manager) diagnoseFailingApps(ctx context.Context, apps []Application) (diag string, terminal bool) {
	if m.kubeClient == nil {
		return "", false
	}
	var b strings.Builder
	for _, app := range apps {
		ns := app.Namespace
		if ns == "" {
			continue
		}
		listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		pods, err := m.kubeClient.CoreV1().Pods(ns).List(listCtx, metav1.ListOptions{})
		cancel()
		if err != nil {
			continue
		}
		header := false
		writeHeader := func() {
			if !header {
				fmt.Fprintf(&b, "\n%s (namespace %s):", app.Name, ns)
				header = true
			}
		}
		for i := range pods.Items {
			for _, ci := range failingContainers(pods.Items[i]) {
				if ci.terminal {
					terminal = true
				}
				writeHeader()
				fmt.Fprintf(&b, "\n  pod %s / %s: %s (restarts=%d)", ci.pod, ci.container, ci.reason, ci.restarts)
				if isImagePullReason(ci.reason) {
					fmt.Fprintf(&b, "\n    image %q could not be pulled — a private image needs registry credentials (an imagePullSecret, or a registry login)", ci.image)
				}
				logs := m.lastPodLogs(ctx, ns, ci.pod, ci.container, ci.restarts > 0)
				if logs == "" && ci.restarts > 0 {
					logs = m.lastPodLogs(ctx, ns, ci.pod, ci.container, false)
				}
				if logs = strings.TrimSpace(logs); logs != "" {
					if len(logs) > diagMaxLogChars {
						logs = "…" + logs[len(logs)-diagMaxLogChars:]
					}
					fmt.Fprintf(&b, "\n    last log lines:\n%s", indentLines(logs, "      "))
				}
			}
		}
		if ev := m.recentWarningEvents(ctx, ns); ev != "" {
			writeHeader()
			fmt.Fprintf(&b, "\n  recent warning events:\n%s", indentLines(ev, "    "))
		}
	}
	return b.String(), terminal
}

// lastPodLogs returns the last few log lines of a container — the PREVIOUS
// (crashed) instance when previous is true (that is where a CrashLoop's fatal
// error is), else the current one. Bounded and best-effort; empty on any error.
func (m *Manager) lastPodLogs(ctx context.Context, ns, pod, container string, previous bool) string {
	if m.kubeClient == nil {
		return ""
	}
	tail := int64(diagLogTailLines)
	opts := &corev1.PodLogOptions{Container: container, TailLines: &tail, Previous: previous}
	logCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	stream, err := m.kubeClient.CoreV1().Pods(ns).GetLogs(pod, opts).Stream(logCtx)
	if err != nil {
		return ""
	}
	defer func() { _ = stream.Close() }()
	var lines []string
	sc := bufio.NewScanner(stream)
	sc.Buffer(make([]byte, 0, 64*1024), 256*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) > diagLogTailLines {
			lines = lines[1:]
		}
	}
	return strings.Join(lines, "\n")
}

// recentWarningEvents returns the most recent Warning events in a namespace
// (Failed image pull, BackOff, unhealthy probes, …), newest first and bounded.
// Best-effort; empty on any error.
func (m *Manager) recentWarningEvents(ctx context.Context, ns string) string {
	if m.kubeClient == nil {
		return ""
	}
	evCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	evs, err := m.kubeClient.CoreV1().Events(ns).List(evCtx, metav1.ListOptions{})
	if err != nil {
		return ""
	}
	type ev struct {
		t time.Time
		s string
	}
	var out []ev
	for _, e := range evs.Items {
		if e.Type != corev1.EventTypeWarning {
			continue
		}
		ts := e.LastTimestamp.Time
		if ts.IsZero() {
			ts = e.EventTime.Time
		}
		out = append(out, ev{ts, fmt.Sprintf("%s %s/%s: %s", e.Reason, e.InvolvedObject.Kind, e.InvolvedObject.Name, strings.TrimSpace(e.Message))})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].t.After(out[j].t) })
	var lines []string
	for i, e := range out {
		if i >= diagMaxEvents {
			break
		}
		lines = append(lines, e.s)
	}
	return strings.Join(lines, "\n")
}

// indentLines prefixes every line of s with prefix.
func indentLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
