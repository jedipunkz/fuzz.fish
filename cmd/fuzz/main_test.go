package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The printed version comes from a linker-stamped variable, so the only honest
// check is to build the binary the way the release workflow does and run it.
func TestVersionFlag(t *testing.T) {
	tests := []struct {
		name    string
		ldflags string
		want    string
	}{
		{name: "default", want: "dev"},
		{name: "stamped", ldflags: "-X main.version=v9.9.9", want: "v9.9.9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bin := filepath.Join(t.TempDir(), "fuzz")
			build := exec.Command("go", "build", "-ldflags", tt.ldflags, "-o", bin, ".")
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("build: %v\n%s", err, out)
			}

			out, err := exec.Command(bin, "--version").Output()
			if err != nil {
				t.Fatalf("run: %v", err)
			}
			if got := strings.TrimSpace(string(out)); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
