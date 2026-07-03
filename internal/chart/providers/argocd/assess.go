package argocd

import (
	"fmt"
	"strings"
)

// appAssessment summarizes one polling tick over the ArgoCD applications.
type appAssessment struct {
	ready        int      // currently Healthy AND Synced
	healthyNames []string // names of currently-Healthy apps
	notReady     []string // "name (status)" labels for apps not yet ready
}

// assessApplications classifies the applications for one polling tick and
// marks apps that are currently Healthy+Synced in everReady (the session-wide
// set of apps that have ever been ready).
func assessApplications(apps []Application, everReady map[string]bool) appAssessment {
	var a appAssessment
	for _, app := range apps {
		if app.Health == "Healthy" {
			a.healthyNames = append(a.healthyNames, app.Name)
		}
		if app.Health == "Healthy" && app.Sync == "Synced" {
			a.ready++
			// Once marked, apps stay counted even if they go out of sync later.
			everReady[app.Name] = true
			continue
		}
		// Show the most important status issue.
		var status string
		switch {
		case app.Health != "Healthy" && app.Sync != "Synced":
			status = fmt.Sprintf("%s/%s", app.Health, app.Sync)
		case app.Health != "Healthy":
			status = fmt.Sprintf("Health: %s", app.Health)
		default:
			status = fmt.Sprintf("Sync: %s", app.Sync)
		}
		a.notReady = append(a.notReady, fmt.Sprintf("%s (%s)", app.Name, status))
	}
	return a
}

// isDeploymentComplete reports whether every currently-visible application is
// ready. The high-water-mark guard withholds completion when the API
// momentarily returns fewer apps than we have ever seen, so success is never
// declared against a partial listing.
func isDeploymentComplete(totalApps, currentlyReady, maxSeenTotal int) bool {
	return totalApps > 0 && currentlyReady == totalApps && totalApps >= maxSeenTotal
}

// repoServerErrorPatterns are condition-message fragments indicating the
// ArgoCD repo-server is failing to serve manifests for an application.
var repoServerErrorPatterns = []string{
	"EOF",
	"Unavailable",
	"error reading from server",
	"failed to generate manifest",
}

// classifyAppIssues splits apps into those stuck in Unknown health/sync and
// those whose condition message points at repo-server trouble. issueCounts is
// updated in place: incremented for apps showing a repo-server error and
// cleared for apps that no longer do.
func classifyAppIssues(apps []Application, issueCounts map[string]int) (unknown, conditionErrors []Application) {
	for _, app := range apps {
		if app.Health == "Unknown" || app.Sync == "Unknown" {
			unknown = append(unknown, app)
		}

		hasRepoErr := false
		if app.Condition != "" {
			for _, p := range repoServerErrorPatterns {
				if strings.Contains(app.Condition, p) {
					hasRepoErr = true
					break
				}
			}
		}
		if hasRepoErr {
			conditionErrors = append(conditionErrors, app)
			issueCounts[app.Name]++
		} else {
			delete(issueCounts, app.Name)
		}
	}
	return unknown, conditionErrors
}
