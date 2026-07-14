package argocd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeValuesFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openframe-helm-values.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing values file: %v", err)
	}
	return path
}

func TestValidateUserValuesFile_MissingFileIsFine(t *testing.T) {
	if err := ValidateUserValuesFile(filepath.Join(t.TempDir(), "nope.yaml")); err != nil {
		t.Fatalf("missing file must validate (baseline install), got: %v", err)
	}
}

func TestValidateUserValuesFile_ValidOverridePasses(t *testing.T) {
	path := writeValuesFile(t, "repository:\n  branch: main\nargocd:\n  dex:\n    enabled: true\n")
	if err := ValidateUserValuesFile(path); err != nil {
		t.Fatalf("valid argocd mapping must pass, got: %v", err)
	}
}

func TestValidateUserValuesFile_NoArgoCDKeyPasses(t *testing.T) {
	path := writeValuesFile(t, "repository:\n  branch: main\n")
	if err := ValidateUserValuesFile(path); err != nil {
		t.Fatalf("file without argocd key must pass, got: %v", err)
	}
}

// TestValidateUserValuesFile_NonMappingArgoCDFails is the pre-flight case from
// the 0.4.9 verification: `argocd: [x]` (or a scalar, or typo'd indentation)
// must fail before any cluster work, naming both the file and the key.
func TestValidateUserValuesFile_NonMappingArgoCDFails(t *testing.T) {
	for name, content := range map[string]string{
		"list":   "argocd:\n  - dex\n",
		"scalar": "argocd: yes\n",
	} {
		t.Run(name, func(t *testing.T) {
			path := writeValuesFile(t, content)
			err := ValidateUserValuesFile(path)
			if err == nil {
				t.Fatal("non-mapping argocd must fail validation")
			}
			if !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), `"argocd"`) {
				t.Errorf("error must name the file and the key, got: %v", err)
			}
		})
	}
}

func TestValidateUserValuesFile_BrokenYAMLFails(t *testing.T) {
	path := writeValuesFile(t, "argocd:\n\tdex: {enabled\n")
	err := ValidateUserValuesFile(path)
	if err == nil {
		t.Fatal("unparseable YAML must fail validation, not silently pass")
	}
	if !strings.Contains(err.Error(), "not valid YAML") {
		t.Errorf("error must say the YAML is invalid, got: %v", err)
	}
}
