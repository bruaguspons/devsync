package mise

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeSnapshot(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mise.toml.snapshot")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write snapshot fixture: %v", err)
	}
	return path
}

func TestDesiredState_PlainStringShape(t *testing.T) {
	path := writeSnapshot(t, `
[tools]
node = "lts"
"aqua:boyter/scc" = "latest"
`)

	p := New(path)
	states, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}

	byName := map[string]string{}
	for _, s := range states {
		byName[s.Name] = s.Version
	}
	if byName["node"] != "lts" {
		t.Errorf("node = %q, want lts", byName["node"])
	}
	if byName["aqua:boyter/scc"] != "latest" {
		t.Errorf("aqua:boyter/scc = %q, want latest", byName["aqua:boyter/scc"])
	}
}

func TestDesiredState_TableShape(t *testing.T) {
	path := writeSnapshot(t, `
[tools.node]
version = "24.0.0"
`)

	p := New(path)
	states, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}
	if len(states) != 1 || states[0].Name != "node" || states[0].Version != "24.0.0" {
		t.Fatalf("unexpected states: %+v", states)
	}
}

func TestDesiredState_ArrayShape(t *testing.T) {
	path := writeSnapshot(t, `
[tools]
node = ["24.0.0", "20.0.0"]
`)

	p := New(path)
	states, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}
	if len(states) != 1 || states[0].Version != "24.0.0" {
		t.Fatalf("expected first array entry to win, got %+v", states)
	}
}

func TestDesiredState_SetsComparisonVersionToRawValue(t *testing.T) {
	path := writeSnapshot(t, `
[tools]
node = "lts"
bun = ["1.3.0", "1.2.0"]

[tools.rust]
version = "latest"
`)

	p := New(path)
	states, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}

	byName := map[string]string{}
	for _, s := range states {
		byName[s.Name] = s.ComparisonVersion
	}
	if byName["node"] != "lts" {
		t.Errorf("node.ComparisonVersion = %q, want lts", byName["node"])
	}
	if byName["rust"] != "latest" {
		t.Errorf("rust.ComparisonVersion = %q, want latest", byName["rust"])
	}
	if byName["bun"] != "1.3.0" {
		t.Errorf("bun.ComparisonVersion = %q, want 1.3.0", byName["bun"])
	}
}

func TestDecodeToolVersion_ArrayOfTablesEmptyVersionRejected(t *testing.T) {
	path := writeSnapshot(t, `
[[tools.node]]
version = ""
`)

	p := New(path)
	_, err := p.DesiredState()
	if err == nil {
		t.Fatal("DesiredState: expected error for empty version in array-of-tables form, got nil")
	}
}

func TestDesiredState_MissingSnapshot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")

	p := New(missing)
	_, err := p.DesiredState()
	if !errors.Is(err, ErrSnapshotMissing) {
		t.Fatalf("expected ErrSnapshotMissing, got %v", err)
	}
}
