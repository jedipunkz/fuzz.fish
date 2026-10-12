package history

import (
	"bytes"
	"testing"
)

func FuzzParseReader(f *testing.F) {
	// Seeds taken from the hand-written cases in parser_test.go.
	f.Add([]byte(""))
	f.Add([]byte("- cmd: ls -la\n  when: 1700000000\n"))
	f.Add([]byte("- cmd: echo hello\n"))
	f.Add([]byte("- cmd: ls\n  when: notanumber\n"))
	f.Add([]byte("- cmd: make build\n  when: 1700000000\n  paths:\n    - /home/user/project\n"))
	f.Add([]byte("- cmd: git status\n  when: 1000\n- cmd: git status\n  when: 2000\n"))
	f.Add([]byte("- cmd: grep '\\\\d' file\n  when: 1000\n- cmd: echo a\\necho b\n  when: 2000\n"))
	f.Add([]byte("- cmd: ls\n  when: 1000\n  paths:\n    - /tmp/back\\\\slash\n"))
	f.Add([]byte("\n- cmd: ls\n  when: 1000\n\n- cmd: pwd\n  when: 2000\n\n"))
	f.Add([]byte("- cmd: \\\\\n  when: 1\n  paths:\n    - \\\\\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		entries := parseReader(bytes.NewReader(data))
		if entries == nil {
			t.Fatal("parseReader returned nil, want an empty slice")
		}
		for _, e := range entries {
			if e.Count < 1 {
				t.Fatalf("entry %q has Count %d, want >= 1", e.Cmd, e.Count)
			}
		}
	})
}
