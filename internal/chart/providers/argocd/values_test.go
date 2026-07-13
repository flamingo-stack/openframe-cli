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
	// dex is intentionally excluded: it is disabled (see TestArgoCDValues_DexDisabled),
	// so it schedules no pod and needs no resource block.
	components := map[string]component{
		"controller":     v.Controller,
		"server":         v.Server,
		"repoServer":     v.RepoServer,
		"redis":          v.Redis,
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

// TestArgoCDValues_DexDisabled locks the V1 blocker fix: the argo-cd chart
// defaults dex.enabled to true, and the dexidp/dex:v2.45.1 arm64 image
// intermittently SIGSEGVs under emulation on Apple Silicon, CrashLoopBackOff-ing
// the 7-minute `helm --wait` into a fresh-install failure. OpenFrame's login
// uses the local developer account, never dex, so it must be disabled.
func TestArgoCDValues_DexDisabled(t *testing.T) {
	var raw struct {
		Dex struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"dex"`
	}
	if err := yaml.Unmarshal([]byte(GetArgoCDValues()), &raw); err != nil {
		t.Fatal(err)
	}
	if raw.Dex.Enabled == nil {
		t.Fatal("dex.enabled must be set explicitly (chart default is true); it is absent")
	}
	if *raw.Dex.Enabled {
		t.Error("dex.enabled must be false — dex is unused and its arm64 image crashes on Apple Silicon")
	}
}
