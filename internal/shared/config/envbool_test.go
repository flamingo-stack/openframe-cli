package config

import "testing"

// TestEnvBool locks strict boolean parsing for OPENFRAME_* switches: the old
// any-non-empty check treated FLAG=0 / FLAG=false as ON — for
// OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY that silently DISABLED release
// signature verification when the user meant "off" (audit B5).
func TestEnvBool(t *testing.T) {
	const name = "OPENFRAME_TEST_BOOL"

	for _, v := range []string{"1", "true", "TRUE", "yes", "on", " 1 "} {
		t.Setenv(name, v)
		if !EnvBool(name) {
			t.Errorf("EnvBool(%q=%q) = false, want true", name, v)
		}
	}
	for _, v := range []string{"", "0", "false", "no", "off", "garbage"} {
		t.Setenv(name, v)
		if EnvBool(name) {
			t.Errorf("EnvBool(%q=%q) = true, want false", name, v)
		}
	}
}
