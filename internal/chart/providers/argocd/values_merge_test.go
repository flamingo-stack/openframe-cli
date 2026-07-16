package argocd

import (
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// parseMerged unmarshals a merged-values YAML string into a generic map.
func parseMerged(t *testing.T, s string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("merged values not valid YAML: %v", err)
	}
	return m
}

// TestMergedArgoCDValues_NoOverridesReturnsBaselineVerbatim: without an
// `argocd:` subtree the baseline is returned byte-for-byte and no keys are
// reported, so the common install path is completely unchanged.
func TestMergedArgoCDValues_NoOverridesReturnsBaselineVerbatim(t *testing.T) {
	for _, uv := range []map[string]interface{}{
		nil,
		{"repository": map[string]interface{}{"branch": "main"}}, // app-of-apps keys, no argocd
		{UserArgoCDKey: map[string]interface{}{}},                // present but empty (argocd: {})
		{UserArgoCDKey: nil},                                     // bare `argocd:` (null)
	} {
		out, keys, err := MergedArgoCDValues(uv)
		if err != nil {
			t.Fatalf("MergedArgoCDValues: %v", err)
		}
		if out != GetArgoCDValues() {
			t.Error("baseline must be returned verbatim when there are no argocd overrides")
		}
		if len(keys) != 0 {
			t.Errorf("no override keys expected, got %v", keys)
		}
	}
}

// TestMergedArgoCDValues_ReEnableDex is the motivating case: a user can flip a
// baseline default back (dex.enabled false -> true) via the file, without a
// flag, and the rest of the baseline (developer account, controller args) must
// survive the merge.
func TestMergedArgoCDValues_ReEnableDex(t *testing.T) {
	uv := map[string]interface{}{
		UserArgoCDKey: map[string]interface{}{
			"dex": map[string]interface{}{"enabled": true},
		},
	}
	out, keys, err := MergedArgoCDValues(uv)
	if err != nil {
		t.Fatalf("MergedArgoCDValues: %v", err)
	}
	if strings.Join(keys, ",") != "dex" {
		t.Errorf("override keys = %v, want [dex]", keys)
	}

	m := parseMerged(t, out)
	dex, _ := m["dex"].(map[string]interface{})
	if dex["enabled"] != true {
		t.Errorf("dex.enabled must be overridden to true, got %v", dex["enabled"])
	}
	// Baseline must survive: the developer bcrypt account and fullnameOverride.
	if m["fullnameOverride"] != "argocd" {
		t.Error("merge dropped fullnameOverride from the baseline")
	}
	configs, _ := m["configs"].(map[string]interface{})
	secret, _ := configs["secret"].(map[string]interface{})
	extra, _ := secret["extra"].(map[string]interface{})
	if pw, _ := extra["accounts.developer.password"].(string); !strings.HasPrefix(pw, "$2a$") {
		t.Error("merge dropped the developer account bcrypt hash from the baseline")
	}
}

// TestMergedArgoCDValues_DeepMergeAndListReplace: nested maps merge (sibling
// keys under a shared parent survive), while a list value replaces wholesale
// (helm semantics).
func TestMergedArgoCDValues_DeepMergeAndListReplace(t *testing.T) {
	uv := map[string]interface{}{
		UserArgoCDKey: map[string]interface{}{
			"configs": map[string]interface{}{
				"params": map[string]interface{}{"server.insecure": "true"}, // new sibling
			},
			"controller": map[string]interface{}{
				"extraArgs": []interface{}{"--custom"}, // replaces the baseline list
			},
		},
	}
	out, _, err := MergedArgoCDValues(uv)
	if err != nil {
		t.Fatalf("MergedArgoCDValues: %v", err)
	}
	m := parseMerged(t, out)
	configs, _ := m["configs"].(map[string]interface{})
	params, _ := configs["params"].(map[string]interface{})
	// New sibling added...
	if params["server.insecure"] != "true" {
		t.Errorf("nested map merge lost the new param, got %v", params["server.insecure"])
	}
	// ...without wiping the baseline siblings under the same parent.
	if params["controller.sync.timeout.seconds"] == nil {
		t.Error("deep merge wiped a baseline sibling param")
	}
	// The list replaced, not appended.
	ctrl, _ := m["controller"].(map[string]interface{})
	args, _ := ctrl["extraArgs"].([]interface{})
	if len(args) != 1 || args[0] != "--custom" {
		t.Errorf("list value must replace wholesale, got %v", args)
	}
}

// TestMergedArgoCDValues_MultipleKeysSorted: the reported override keys are
// sorted (stable, greppable warning output).
func TestMergedArgoCDValues_MultipleKeysSorted(t *testing.T) {
	uv := map[string]interface{}{
		UserArgoCDKey: map[string]interface{}{
			"server":     map[string]interface{}{"replicas": 2},
			"dex":        map[string]interface{}{"enabled": true},
			"controller": map[string]interface{}{"replicas": 1},
		},
	}
	_, keys, err := MergedArgoCDValues(uv)
	if err != nil {
		t.Fatalf("MergedArgoCDValues: %v", err)
	}
	if strings.Join(keys, ",") != "controller,dex,server" {
		t.Errorf("keys must be sorted, got %v", keys)
	}
}

// TestDeepMerge_Semantics unit-tests the merge rule directly.
func TestDeepMerge_Semantics(t *testing.T) {
	dst := map[string]interface{}{
		"keep":   "baseline",
		"nested": map[string]interface{}{"a": 1, "b": 2},
		"list":   []interface{}{"x", "y"},
	}
	src := map[string]interface{}{
		"nested": map[string]interface{}{"b": 20, "c": 3}, // merge: a survives, b replaced, c added
		"list":   []interface{}{"z"},                      // replace
		"new":    "added",
	}
	deepMerge(dst, src)

	if dst["keep"] != "baseline" {
		t.Error("untouched key must survive")
	}
	n := dst["nested"].(map[string]interface{})
	if n["a"] != 1 || n["b"] != 20 || n["c"] != 3 {
		t.Errorf("nested merge wrong: %v", n)
	}
	if l := dst["list"].([]interface{}); len(l) != 1 || l[0] != "z" {
		t.Errorf("list must replace, got %v", l)
	}
	if dst["new"] != "added" {
		t.Error("new key must be added")
	}
}

// TestMergedArgoCDValues_MalformedOverrideErrors: a present-but-non-map argocd:
// value (scalar, list, typo'd indentation) is a mistake, not a no-op. It must
// error so installArgoCDHelm surfaces it, rather than silently dropping the
// user's intended override (the V3 silent-failure class).
func TestMergedArgoCDValues_MalformedOverrideErrors(t *testing.T) {
	cases := map[string]interface{}{
		"scalar bool":   true,
		"scalar string": "enabled",
		"list":          []interface{}{"dex"},
		"number":        42,
	}
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			out, keys, err := MergedArgoCDValues(map[string]interface{}{UserArgoCDKey: bad})
			if err == nil {
				t.Fatalf("a malformed %q override must error, got out=%q keys=%v", UserArgoCDKey, out, keys)
			}
			if !strings.Contains(err.Error(), UserArgoCDKey) || !strings.Contains(err.Error(), "mapping") {
				t.Errorf("error must name the key and say it must be a mapping, got: %v", err)
			}
			if out != "" {
				t.Errorf("on error the values string must be empty, got %q", out)
			}
		})
	}
}
