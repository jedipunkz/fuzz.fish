package config

import "testing"

func TestParseKeymap_Defaults(t *testing.T) {
	km, err := ParseKeymap(nil)
	if err != nil {
		t.Fatal(err)
	}
	for action, keys := range DefaultKeybinds {
		for _, key := range keys {
			if km[key] != action {
				t.Errorf("km[%q] = %q, want %q", key, km[key], action)
			}
		}
	}
}

func TestParseKeymap_Override(t *testing.T) {
	km, err := ParseKeymap([]byte("keybinds:\n  files: [ctrl+f]\n  copy: []\n"))
	if err != nil {
		t.Fatal(err)
	}
	if km["ctrl+f"] != ActionFiles {
		t.Errorf("ctrl+f = %q, want files", km["ctrl+f"])
	}
	if _, ok := km["ctrl+s"]; ok {
		t.Error("ctrl+s should no longer be bound")
	}
	if _, ok := km["ctrl+y"]; ok {
		t.Error("ctrl+y should be unbound")
	}
	if km["ctrl+r"] != ActionHistory {
		t.Errorf("ctrl+r = %q, want history (default kept)", km["ctrl+r"])
	}
}

func TestParseKeymap_Errors(t *testing.T) {
	for name, in := range map[string]string{
		"unknown action": "keybinds:\n  nope: [ctrl+a]\n",
		"unknown field":  "keybind:\n  files: [ctrl+f]\n",
		"duplicate key":  "keybinds:\n  files: [ctrl+r]\n",
	} {
		if _, err := ParseKeymap([]byte(in)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
