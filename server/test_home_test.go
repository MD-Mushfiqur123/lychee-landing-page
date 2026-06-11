package server

import (
	"testing"

	"github.com/lychee/lychee/envconfig"
)

func setTestHome(t *testing.T, home string) {
	t.Helper()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LYCHEE_MODELS", "")
	envconfig.ReloadServerConfig()
}
