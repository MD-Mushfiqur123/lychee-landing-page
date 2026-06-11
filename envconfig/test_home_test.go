package envconfig

import "testing"

func setTestHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LYCHEE_MODELS", "")
	ReloadServerConfig()
}
