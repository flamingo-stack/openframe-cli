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
// waiting with an error reason (excluding the benign startup reasons),
// terminated with a non-zero exit, or crash-looping while momentarily Running
// between restarts. Pure and testable.
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
		case cs.RestartCount >= crashLoopRestartThreshold &&
			cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.ExitCode != 0:
			// A CrashLooping container spends part of each cycle Running (or in a
			// benign waiting state), where only the restart count and the previous
			// instance's non-zero exit betray the loop — without this branch such
			// a pod is skipped and its crash logs never pulled.
			reason = "CrashLoop (running between restarts)"
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

// appPods lists the pods belonging to an ArgoCD application: first by ArgoCD's
// default tracking label, then by its configurable alternative. owned reports
// whether the returned pods are attributable to the app; when neither selector
// matches anything, it falls back to the whole namespace — still useful for
// the human diagnostic, but NOT safe to base a fail-fast decision on: an
// app-of-apps platform routinely shares destination namespaces, so a leftover
// CrashLooping pod of a neighbouring app must not abort THIS app's install.
func (m *Manager) appPods(ctx context.Context, ns, appName string) (pods []corev1.Pod, owned bool) {
	for _, sel := range []string{"app.kubernetes.io/instance=" + appName, "argocd.argoproj.io/instance=" + appName} {
		listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		list, err := m.kubeClient.CoreV1().Pods(ns).List(listCtx, metav1.ListOptions{LabelSelector: sel})
		cancel()
		if err == nil && len(list.Items) > 0 {
			return list.Items, true
		}
	}
	listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	list, err := m.kubeClient.CoreV1().Pods(ns).List(listCtx, metav1.ListOptions{})
	if err != nil {
		return nil, false
	}
	return list.Items, false
}

// diagnoseFailingApps inspects the workloads of the given apps and returns a
// human diagnostic — the failing pod/container, why (waiting reason or crash), a
// registry-auth hint for image-pull failures, the crashed container's last log
// lines, and recent warning events. stuck lists the apps whose OWN workloads
// are in a won't-recover state (CrashLoop / ImagePull / repeated crashes),
// letting the caller fail fast instead of waiting out the timeout — and only
// for those apps, not every candidate sharing the namespace. Best-effort.
func (m *Manager) diagnoseFailingApps(ctx context.Context, apps []Application) (diag string, stuck []Application) {
	if m.kubeClient == nil {
		return "", nil
	}
	var b strings.Builder
	for _, app := range apps {
		ns := app.Namespace
		if ns == "" {
			continue
		}
		pods, owned := m.appPods(ctx, ns, app.Name)
		header := false
		writeHeader := func() {
			if !header {
				fmt.Fprintf(&b, "\n%s (namespace %s):", app.Name, ns)
				header = true
			}
		}
		appTerminal := false
		for i := range pods {
			for _, ci := range failingContainers(pods[i]) {
				if ci.terminal {
					appTerminal = true
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
		// Only attributable failures may abort: a terminal pod found via the
		// namespace fallback could belong to any app (or to nothing at all).
		if appTerminal && owned {
			stuck = append(stuck, app)
		}
	}
	return b.String(), stuck
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
