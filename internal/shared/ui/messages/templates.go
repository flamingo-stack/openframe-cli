package messages

// NOTE: This package uses a template system where formatting arguments are passed to
// Show* methods which then format them through FormatMessage(). The go vet tool
// doesn't understand this indirection and reports "no formatting directives" warnings.
// These warnings are false positives - the formatting is handled correctly by the
// template system. The functionality works as intended.

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/pterm/pterm"
)

// MessageType represents different types of messages
type MessageType int

const (
	InfoMessage MessageType = iota
	SuccessMessage
	WarningMessage
	ErrorMessage
	ProgressMessage
	CompletionMessage
)

// Templates provides standardized message templates
type Templates struct {
	templates map[MessageType]map[string]string
}

// NewTemplates creates a new message templates instance
func NewTemplates() *Templates {
	return &Templates{
		templates: map[MessageType]map[string]string{
			InfoMessage: {
				"operation_start":    "📦 Starting %s on cluster: %s",
				"operation_progress": "🔄 %s - %s",
				"checking":           "🔍 %s",
				"connecting":         "🔗 Connecting to %s",
				"downloading":        "📥 Downloading %s",
				"installing":         "⚙️  Installing %s",
				"configuring":        "⚙️  Configuring %s",
				"validating":         "✔️  Validating %s",
				"waiting":            "⏳ Waiting for %s",
				"next_steps":         "🚀 Next Steps:",
			},
			SuccessMessage: {
				"operation_complete":     "✅ %s completed successfully!",
				"installation_complete":  "✅ %s installation completed successfully!",
				"step_complete":          "✅ %s completed (%s)",
				"credentials_provided":   "✅ Credentials provided",
				"validation_passed":      "✅ %s validation passed",
				"connection_established": "✅ Connection to %s established",
			},
			WarningMessage: {
				"operation_cancelled": "No %s selected. %s cancelled.",
				"step_skipped":        "⏭️  %s skipped: %s",
				"partial_success":     "⚠️  %s completed with warnings",
				"deprecated_feature":  "⚠️  %s is deprecated, consider using %s instead",
			},
			ErrorMessage: {
				"operation_failed":    "❌ %s failed: %v",
				"step_failed":         "❌ %s failed: %v (%s)",
				"installation_failed": "❌ %s installation failed: %v",
				"validation_failed":   "❌ %s validation failed: %v",
				"connection_failed":   "❌ Failed to connect to %s: %v",
				"not_found":           "❌ %s '%s' not found",
				"invalid_input":       "❌ Invalid %s: %s",
				"permission_denied":   "❌ Permission denied for %s: %v",
				"timeout":             "❌ %s timed out after %s",
			},
			ProgressMessage: {
				"bootstrapping":  "⏳ Waiting %s for %s to bootstrap...",
				"health_check":   "⏳ Waiting for %s to be healthy and ready...",
				"sync_check":     "⏳ Waiting for %s to sync...",
				"resource_ready": "⏳ Waiting for %s resources to be ready...",
			},
			CompletionMessage: {
				"troubleshooting": "🔧 Troubleshooting steps:",
				"summary":         "📊 %s Summary:",
				"duration":        "Total Duration: %s",
				"statistics":      "Steps: %d completed, %d failed, %d skipped, %d total",
			},
		},
	}
}

// FormatMessage formats a message using the specified template.
//
// NOTE: the second parameter is a template *key* (looked up in t.templates),
// not a printf format string. The fallback below deliberately does NOT pass
// `template` as a format string to fmt.Sprintf — doing so makes `go vet`'s
// printf analyzer infer this function (and every Show*/render* wrapper that
// forwards to it) as a printf wrapper, producing false-positive
// "no formatting directives" warnings at all call sites. Keeping the dynamic
// format string in the `format` local (the map value) confines printf
// inference to that single, legitimate Sprintf call.
func (t *Templates) FormatMessage(msgType MessageType, template string, args ...interface{}) string {
	if templates, exists := t.templates[msgType]; exists {
		if format, exists := templates[template]; exists {
			return fmt.Sprintf(format, args...)
		}
	}
	// Fallback: template key not found. Render the key followed by any args
	// without using the key itself as a format string.
	if len(args) == 0 {
		return template
	}
	return strings.TrimRight(template+": "+fmt.Sprintln(args...), "\n")
}

// capitalizeFirst upper-cases the first rune of a single-word label.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// renderMessage is a generic message renderer that bypasses go vet checks
//
//nolint:govet
func (t *Templates) renderMessage(msgType MessageType, template string, args ...interface{}) {
	message := t.FormatMessage(msgType, template, args...)

	switch msgType {
	case InfoMessage:
		pterm.Info.Println(message)
	case SuccessMessage:
		pterm.Success.Println(message)
	case WarningMessage:
		pterm.Warning.Println(message)
	case ErrorMessage:
		pterm.Error.Println(message)
	case ProgressMessage:
		pterm.Info.Println(message)
	default:
		pterm.Println(message)
	}
}

// renderInfo renders and displays an info message
func (t *Templates) renderInfo(template string, args ...interface{}) {
	t.renderMessage(InfoMessage, template, args...) //nolint:govet
}

// renderSuccess renders and displays a success message
func (t *Templates) renderSuccess(template string, args ...interface{}) {
	t.renderMessage(SuccessMessage, template, args...) //nolint:govet
}

// renderWarning renders and displays a warning message
func (t *Templates) renderWarning(template string, args ...interface{}) {
	t.renderMessage(WarningMessage, template, args...) //nolint:govet
}

// renderError renders and displays an error message
func (t *Templates) renderError(template string, args ...interface{}) {
	t.renderMessage(ErrorMessage, template, args...) //nolint:govet
}

// renderProgress renders and displays a progress message
func (t *Templates) renderProgress(template string, args ...interface{}) {
	t.renderMessage(ProgressMessage, template, args...) //nolint:govet
}

// ShowInfo displays an info message using templates
func (t *Templates) ShowInfo(template string, args ...interface{}) {
	t.renderInfo(template, args...)
}

// ShowSuccess displays a success message using templates
func (t *Templates) ShowSuccess(template string, args ...interface{}) {
	t.renderSuccess(template, args...)
}

// ShowWarning displays a warning message using templates
func (t *Templates) ShowWarning(template string, args ...interface{}) {
	t.renderWarning(template, args...)
}

// ShowError displays an error message using templates
func (t *Templates) ShowError(template string, args ...interface{}) {
	t.renderError(template, args...)
}

// ShowProgress displays a progress message using templates
func (t *Templates) ShowProgress(template string, args ...interface{}) {
	t.renderProgress(template, args...)
}

// ShowOperationStart shows a standardized operation start message
func (t *Templates) ShowOperationStart(operation, target string) {
	t.renderMessage(InfoMessage, "operation_start", operation, pterm.Cyan(target)) //nolint:govet
}

// ShowOperationComplete shows a standardized operation completion message
func (t *Templates) ShowOperationComplete(operation string) {
	t.renderMessage(SuccessMessage, "operation_complete", operation) //nolint:govet
}

// ShowOperationFailed shows a standardized operation failure message
func (t *Templates) ShowOperationFailed(operation string, err error) {
	t.renderMessage(ErrorMessage, "operation_failed", operation, err) //nolint:govet
}

// ShowStepComplete shows a standardized step completion message
func (t *Templates) ShowStepComplete(stepName string, duration time.Duration) {
	t.renderMessage(SuccessMessage, "step_complete", stepName, duration.Round(time.Millisecond)) //nolint:govet
}

// ShowStepFailed shows a standardized step failure message
func (t *Templates) ShowStepFailed(stepName string, err error, duration time.Duration) {
	t.renderMessage(ErrorMessage, "step_failed", stepName, err, duration.Round(time.Millisecond)) //nolint:govet
}

// ShowInstallationComplete shows completion message with next steps
func (t *Templates) ShowInstallationComplete(component string, nextSteps []string) {
	t.renderMessage(SuccessMessage, "installation_complete", component) //nolint:govet
	fmt.Println()

	if len(nextSteps) > 0 {
		t.renderInfo("next_steps")
		for i, step := range nextSteps {
			pterm.Printf("  %d. %s\n", i+1, step)
		}
	}
}

// ShowTroubleshootingSteps shows standardized troubleshooting information
func (t *Templates) ShowTroubleshootingSteps(steps []string) {
	fmt.Println()
	pterm.Info.Println("🔧 Troubleshooting steps:")
	for i, step := range steps {
		pterm.Printf("  %d. %s\n", i+1, step)
	}
}

// ShowResourceNotFound shows a standardized not found message
func (t *Templates) ShowResourceNotFound(resourceType, resourceName string) {
	t.renderMessage(ErrorMessage, "not_found", resourceType, resourceName) //nolint:govet
}

// ShowOperationCancelled shows a standardized cancellation message
func (t *Templates) ShowOperationCancelled(resource, operation string) {
	t.renderMessage(WarningMessage, "operation_cancelled", resource, capitalizeFirst(operation)) //nolint:govet
}

// ShowValidationError shows a standardized validation error
func (t *Templates) ShowValidationError(field, reason string) {
	t.renderMessage(ErrorMessage, "invalid_input", field, reason) //nolint:govet
}

// ShowConnectionStatus shows connection status messages
func (t *Templates) ShowConnectionStatus(service string, success bool, err error) {
	if success {
		t.renderMessage(SuccessMessage, "connection_established", service) //nolint:govet
	} else {
		t.renderMessage(ErrorMessage, "connection_failed", service, err) //nolint:govet
	}
}

// ShowBootstrapWait shows bootstrap waiting message
func (t *Templates) ShowBootstrapWait(duration string, service string) {
	t.renderMessage(ProgressMessage, "bootstrapping", duration, service) //nolint:govet
}

// ShowHealthCheck shows health check waiting message
func (t *Templates) ShowHealthCheck(service string) {
	t.renderMessage(ProgressMessage, "health_check", service) //nolint:govet
}

// CustomTemplates allows adding custom message templates
type CustomTemplates struct {
	*Templates
	custom map[MessageType]map[string]string
}

// NewCustomTemplates creates templates with custom additions
func NewCustomTemplates() *CustomTemplates {
	return &CustomTemplates{
		Templates: NewTemplates(),
		custom:    make(map[MessageType]map[string]string),
	}
}

// AddTemplate adds a custom template
func (ct *CustomTemplates) AddTemplate(msgType MessageType, name, format string) {
	if ct.custom[msgType] == nil {
		ct.custom[msgType] = make(map[string]string)
	}
	ct.custom[msgType][name] = format
}

// FormatMessage overrides to check custom templates first
func (ct *CustomTemplates) FormatMessage(msgType MessageType, template string, args ...interface{}) string {
	// Check custom templates first
	if customs, exists := ct.custom[msgType]; exists {
		if format, exists := customs[template]; exists {
			return fmt.Sprintf(format, args...)
		}
	}

	// Fall back to standard templates
	return ct.Templates.FormatMessage(msgType, template, args...)
}

// Formatter provides quick access to commonly used formatting functions
type Formatter struct {
	templates *Templates
}

// NewFormatter creates a new message formatter
func NewFormatter() *Formatter {
	return &Formatter{
		templates: NewTemplates(),
	}
}

// Installation provides installation-specific message formatting
func (f *Formatter) Installation() *InstallationFormatter {
	return &InstallationFormatter{f.templates}
}

// Cluster provides cluster-specific message formatting
func (f *Formatter) Cluster() *ClusterFormatter {
	return &ClusterFormatter{f.templates}
}

// InstallationFormatter provides installation-specific messages
type InstallationFormatter struct {
	templates *Templates
}

// Starting shows installation start message
func (f *InstallationFormatter) Starting(component, cluster string) {
	f.templates.ShowOperationStart("installation", fmt.Sprintf("%s on %s", component, cluster))
}

// Complete shows installation completion with next steps
func (f *InstallationFormatter) Complete(component string, nextSteps []string) {
	f.templates.ShowInstallationComplete(component, nextSteps)
}

// Failed shows installation failure with troubleshooting
func (f *InstallationFormatter) Failed(component string, err error, troubleshootingSteps []string) {
	f.templates.ShowOperationFailed(fmt.Sprintf("%s installation", component), err)
	if len(troubleshootingSteps) > 0 {
		f.templates.ShowTroubleshootingSteps(troubleshootingSteps)
	}
}

// ClusterFormatter provides cluster-specific messages
type ClusterFormatter struct {
	templates *Templates
}

// NotFound shows cluster not found message
func (f *ClusterFormatter) NotFound(clusterName string) {
	f.templates.ShowResourceNotFound("cluster", clusterName)
}

// SelectionCancelled shows cluster selection cancelled message
func (f *ClusterFormatter) SelectionCancelled(operation string) {
	f.templates.ShowOperationCancelled("cluster", operation)
}
