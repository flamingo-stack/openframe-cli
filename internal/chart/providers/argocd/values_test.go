package argocd

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type component struct {
	Resources struct {
		Requests map[string]string `yaml:"requests"`
		Limits   map[string]string `yaml:"limits"`
	} `yaml:"resources"`
}

type argoValues struct {
	FullnameOverride string `yaml:"fullnameOverride"`
	Configs          struct {
		Cm     map[string]any `yaml:"cm"`
		Params map[string]any `yaml:"params"`
		Secret struct {
			Extra map[string]any `yaml:"extra"`
		} `yaml:"secret"`
		Rbac map[string]any `yaml:"rbac"`
	} `yaml:"configs"`
	Controller     component `yaml:"controller"`
	Server         component `yaml:"server"`
	RepoServer     component `yaml:"repoServer"`
	Redis          component `yaml:"redis"`
	Dex            component `yaml:"dex"`
	ApplicationSet component `yaml:"applicationSet"`
	Notifications  component `yaml:"notifications"`
}

func parseValues(t *testing.T) argoValues {
	t.Helper()
	var v argoValues
	// Real unmarshal (not substring matching) — catches broken indentation or a
	// malformed Lua/jq block before it ever reaches helm.
	if err := yaml.Unmarshal([]byte(GetArgoCDValues()), &v); err != nil {
		t.Fatalf("embedded argocd-values.yaml is not valid YAML: %v", err)
	}
	return v
}

func TestArgoCDValues_ParsesAndCoreInvariants(t *testing.T) {
	v := parseValues(t)

	// The install and namespace logic depend on the release being named "argocd".
	if v.FullnameOverride != "argocd" {
		t.Errorf("fullnameOverride = %q, want argocd", v.FullnameOverride)
	}
}

func TestArgoCDValues_Params(t *testing.T) {
	v := parseValues(t)
	want := map[string]string{
		"controller.sync.timeout.seconds":                   "1800",
		"applicationsetcontroller.enable.progressive.syncs": "true",
	}
	for k, exp := range want {
		if got := v.Configs.Params[k]; got != exp {
			t.Errorf("configs.params[%q] = %v, want %q", k, got, exp)
		}
	}
}

func TestArgoCDValues_ConfigsCustomizations(t *testing.T) {
	v := parseValues(t)
	for _, key := range []string{
		"resource.customizations.health.argoproj.io_Application",
		"resource.customizations.ignoreDifferences.argoproj.io_Application",
		"accounts.developer",
	} {
		if _, ok := v.Configs.Cm[key]; !ok {
			t.Errorf("configs.cm missing key %q", key)
		}
	}
	// Free-form blocks: confirm the health Lua and the finalizer jq survived.
	if s, _ := v.Configs.Cm["resource.customizations.health.argoproj.io_Application"].(string); !strings.Contains(s, "if obj.status ~= nil then") {
		t.Error("health customization lost its Lua body")
	}
	if s, _ := v.Configs.Cm["resource.customizations.ignoreDifferences.argoproj.io_Application"].(string); !strings.Contains(s, "pre-delete-finalizer") {
		t.Error("ignoreDifferences lost its jq finalizer expression")
	}
}

func TestArgoCDValues_DeveloperAccountDefault(t *testing.T) {
	v := parseValues(t)

	pw, ok := v.Configs.Secret.Extra["accounts.developer.password"].(string)
	if !ok || !strings.HasPrefix(pw, "$2a$") {
		t.Errorf("developer password must be a bcrypt hash, got %q", pw)
	}
	rbac, _ := v.Configs.Rbac["policy.csv"].(string)
	if !strings.Contains(rbac, "role:developer") {
		t.Error("rbac policy.csv missing the developer role")
	}
}

func TestArgoCDValues_ControllerExtraArgs(t *testing.T) {
	var raw struct {
		Controller struct {
			ExtraArgs []string `yaml:"extraArgs"`
		} `yaml:"controller"`
	}
	if err := yaml.Unmarshal([]byte(GetArgoCDValues()), &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.Controller.ExtraArgs) == 0 {
		t.Error("controller.extraArgs must not be empty")
	}
}

// TestArgoCDValues_AllComponentsHaveResources guards that every ArgoCD
// component ships explicit resource requests+limits (cpu+memory) — required for
// scheduling on the small k3d clusters the CLI targets.
func TestArgoCDValues_AllComponentsHaveResources(t *testing.T) {
	v := parseValues(t)
	components := map[string]component{
		"controller":     v.Controller,
		"server":         v.Server,
		"repoServer":     v.RepoServer,
		"redis":          v.Redis,
		"dex":            v.Dex,
		"applicationSet": v.ApplicationSet,
		"notifications":  v.Notifications,
	}
	for name, c := range components {
		for _, res := range []string{"cpu", "memory"} {
			if c.Resources.Requests[res] == "" {
				t.Errorf("%s: missing resources.requests.%s", name, res)
			}
			if c.Resources.Limits[res] == "" {
				t.Errorf("%s: missing resources.limits.%s", name, res)
			}
		}
	}
}
