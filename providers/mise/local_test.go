package mise

import (
	"errors"
	"os"
	"testing"

	"github.com/bruaguspons/devsync/core"
)

func TestLocalState_MiseNotInstalled(t *testing.T) {
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)

	p := &Provider{}
	_, err := p.LocalState()
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("LocalState() error = %v, want errors.Is(err, ErrNotInstalled)", err)
	}
}

func TestDecodeMiseList_RealFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/mise_ls_sample.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	states, err := decodeMiseList(data)
	if err != nil {
		t.Fatalf("decodeMiseList: %v", err)
	}

	byName := map[string]core.Resource{}
	for _, s := range states {
		byName[s.Name] = s
	}

	if len(states) != 11 {
		t.Fatalf("expected 11 tools, got %d: %+v", len(states), states)
	}

	if got, ok := byName["aqua:boyter/scc"]; !ok || got.Version != "3.7.0" || got.Kind != core.ResourceKindTool {
		t.Errorf("aqua:boyter/scc = %+v, want version 3.7.0, kind tool", got)
	}
	if got, ok := byName["node"]; !ok || got.Version != "24.19.0" {
		t.Errorf("node = %+v, want version 24.19.0", got)
	}
	// rust has an extra "symlinked_to" field in the real payload;
	// decode must tolerate unknown fields without error.
	if got, ok := byName["rust"]; !ok || got.Version != "1.97.1" {
		t.Errorf("rust = %+v, want version 1.97.1", got)
	}
}

func TestDecodeMiseList_SetsComparisonVersionFromRequestedVersion(t *testing.T) {
	data, err := os.ReadFile("testdata/mise_ls_sample.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	states, err := decodeMiseList(data)
	if err != nil {
		t.Fatalf("decodeMiseList: %v", err)
	}

	byName := map[string]core.Resource{}
	for _, s := range states {
		byName[s.Name] = s
	}

	if got, ok := byName["node"]; !ok || got.ComparisonVersion != "lts" {
		t.Errorf("node.ComparisonVersion = %q, want %q", got.ComparisonVersion, "lts")
	}
	if got, ok := byName["rust"]; !ok || got.ComparisonVersion != "latest" {
		t.Errorf("rust.ComparisonVersion = %q, want %q", got.ComparisonVersion, "latest")
	}
}

func TestDecodeMiseList_PicksActiveVersion(t *testing.T) {
	data := []byte(`{
		"node": [
			{"version": "20.0.0", "requested_version": "20", "installed": true, "active": false},
			{"version": "24.19.0", "requested_version": "lts", "installed": true, "active": true}
		]
	}`)

	states, err := decodeMiseList(data)
	if err != nil {
		t.Fatalf("decodeMiseList: %v", err)
	}
	if len(states) != 1 || states[0].Version != "24.19.0" {
		t.Fatalf("expected active version 24.19.0, got %+v", states)
	}
}

func TestDecodeMiseList_MalformedJSON(t *testing.T) {
	_, err := decodeMiseList([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error decoding malformed JSON, got nil")
	}
}

func TestNormalizeToolName_PreservesBackendPrefix(t *testing.T) {
	got := normalizeToolName(`"aqua:boyter/scc"`)
	want := "aqua:boyter/scc"
	if got != want {
		t.Errorf("normalizeToolName() = %q, want %q", got, want)
	}
}
