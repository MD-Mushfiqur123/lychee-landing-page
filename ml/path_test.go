package ml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLibLycheePath(t *testing.T) {
	root := t.TempDir()

	tests := []struct {
		name   string
		search libLycheePathSearch
		dirs   []string
		want   string
	}{
		{
			name: "darwin release layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "darwin-app", "Lychee.app", "Contents", "Resources", "lychee"),
				goos:       "darwin",
				goarch:     "arm64",
			},
			dirs: []string{filepath.Join(root, "darwin-app", "Lychee.app", "Contents", "Resources")},
			want: filepath.Join(root, "darwin-app", "Lychee.app", "Contents", "Resources"),
		},
		{
			name: "darwin standard install layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "darwin-install", "bin", "lychee"),
				goos:       "darwin",
				goarch:     "arm64",
			},
			dirs: []string{filepath.Join(root, "darwin-install", "lib", "lychee")},
			want: filepath.Join(root, "darwin-install", "lib", "lychee"),
		},
		{
			name: "windows release layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "windows-release", "lychee.exe"),
				goos:       "windows",
				goarch:     "amd64",
			},
			dirs: []string{filepath.Join(root, "windows-release", "lib", "lychee")},
			want: filepath.Join(root, "windows-release", "lib", "lychee"),
		},
		{
			name: "windows standard install layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "windows-install", "bin", "lychee.exe"),
				goos:       "windows",
				goarch:     "amd64",
			},
			dirs: []string{filepath.Join(root, "windows-install", "lib", "lychee")},
			want: filepath.Join(root, "windows-install", "lib", "lychee"),
		},
		{
			name: "linux standard install layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "linux-install", "bin", "lychee"),
				goos:       "linux",
				goarch:     "amd64",
			},
			dirs: []string{filepath.Join(root, "linux-install", "lib", "lychee")},
			want: filepath.Join(root, "linux-install", "lib", "lychee"),
		},
		{
			name: "local linux underscore dist layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "linux-dev", "lychee"),
				workingDir: filepath.Join(root, "linux-dev"),
				goos:       "linux",
				goarch:     "amd64",
			},
			dirs: []string{filepath.Join(root, "linux-dev", "dist", "linux_amd64", "lib", "lychee")},
			want: filepath.Join(root, "linux-dev", "dist", "linux_amd64", "lib", "lychee"),
		},
		{
			name: "mlx-only standard install layout",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "mlx-install", "bin", "lychee"),
				goos:       "linux",
				goarch:     "amd64",
			},
			dirs: []string{filepath.Join(root, "mlx-install", "lib", "lychee")},
			want: filepath.Join(root, "mlx-install", "lib", "lychee"),
		},
		{
			name: "darwin local build layout before executable directory fallback",
			search: libLycheePathSearch{
				executable: filepath.Join(root, "darwin-dev", "lychee"),
				workingDir: filepath.Join(root, "darwin-dev"),
				goos:       "darwin",
				goarch:     "arm64",
			},
			dirs: []string{filepath.Join(root, "darwin-dev", "build", "lib", "lychee")},
			want: filepath.Join(root, "darwin-dev", "build", "lib", "lychee"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			got := findLibLycheePath(tt.search)
			if got != tt.want {
				t.Fatalf("findLibLycheePath() = %q, want %q; candidates: %v", got, tt.want, libLycheePathCandidates(tt.search))
			}
		})
	}
}
