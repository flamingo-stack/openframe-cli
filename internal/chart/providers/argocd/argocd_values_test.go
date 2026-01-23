package argocd

import (
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/chart/models"
)

func TestGetArgoCDValues(t *testing.T) {
	values := GetArgoCDValues(nil)

	// Test that the function returns non-empty string
	if values == "" {
		t.Error("GetArgoCDValues() returned empty string")
	}

	// Test that it contains expected YAML content
	expectedContent := []string{
		"fullnameOverride: argocd",
		"configs:",
		"resource.customizations.health.argoproj.io_Application:",
		"hs.status = \"Progressing\"",
		"controller.sync.timeout.seconds:",
		"controller:",
		"server:",
		"repoServer:",
		"redis:",
		"dex:",
		"applicationSet:",
		"notifications:",
		"resources:",
		"cpu:",
		"memory:",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(values, expected) {
			t.Errorf("GetArgoCDValues() missing expected content: %s", expected)
		}
	}

	// Test that it's valid YAML format (starts with valid YAML)
	if !strings.Contains(values, "fullnameOverride:") {
		t.Error("GetArgoCDValues() does not appear to be valid YAML format")
	}
}

func TestGetArgoCDValuesStructure(t *testing.T) {
	values := GetArgoCDValues(nil)

	// Count lines to ensure we have the expected structure
	lines := strings.Split(values, "\n")
	if len(lines) < 80 {
		t.Errorf("GetArgoCDValues() returned too few lines: got %d, want at least 80", len(lines))
	}

	// Check for health check script presence
	if !strings.Contains(values, "if obj.status ~= nil then") {
		t.Error("GetArgoCDValues() missing Lua health check script")
	}
}

func TestGetArgoCDValuesWithDefaults(t *testing.T) {
	values := GetArgoCDValues(nil)

	// Check default image repositories are present
	expectedDefaults := []string{
		"repository: ghcr.io/flamingo-stack/registry/argoproj/argocd",
		"tag: v3.2.5",
		"repository: ghcr.io/flamingo-stack/registry/redis",
		"tag: 8.2.2-alpine",
		"repository: ghcr.io/flamingo-stack/registry/dexidp/dex",
		"tag: v2.44.0",
	}

	for _, expected := range expectedDefaults {
		if !strings.Contains(values, expected) {
			t.Errorf("GetArgoCDValues(nil) missing default: %s", expected)
		}
	}
}

func TestGetArgoCDValuesWithCustomConfig(t *testing.T) {
	config := &models.ArgoCDConfig{
		Image: models.ArgoCDImageConfig{
			Repository: "custom-registry/argocd",
			Tag:        "v2.0.0",
		},
		Redis: models.ArgoCDImageConfig{
			Repository: "custom-registry/redis",
			Tag:        "7.0.0",
		},
	}

	values := GetArgoCDValues(config)

	// Check custom values are present
	if !strings.Contains(values, "repository: custom-registry/argocd") {
		t.Error("GetArgoCDValues() did not use custom argocd repository")
	}
	if !strings.Contains(values, "tag: v2.0.0") {
		t.Error("GetArgoCDValues() did not use custom argocd tag")
	}
	if !strings.Contains(values, "repository: custom-registry/redis") {
		t.Error("GetArgoCDValues() did not use custom redis repository")
	}

	// Check that defaults are used for non-overridden values
	if !strings.Contains(values, "repository: ghcr.io/flamingo-stack/registry/dexidp/dex") {
		t.Error("GetArgoCDValues() did not preserve default dex repository")
	}
}

func TestGetArgoCDValuesPartialOverride(t *testing.T) {
	// Only override tag, keep default repository
	config := &models.ArgoCDConfig{
		Image: models.ArgoCDImageConfig{
			Tag: "v3.0.0",
		},
	}

	values := GetArgoCDValues(config)

	// Check that repository defaults are preserved but tag is overridden
	if !strings.Contains(values, "repository: ghcr.io/flamingo-stack/registry/argoproj/argocd") {
		t.Error("GetArgoCDValues() did not preserve default argocd repository")
	}
	if !strings.Contains(values, "tag: v3.0.0") {
		t.Error("GetArgoCDValues() did not use custom argocd tag")
	}
}
