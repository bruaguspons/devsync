package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bruaguspons/devsync/core"
)

func TestLocalState_WalksInstalledSkills(t *testing.T) {
	installedRoot := t.TempDir()
	writeFile(t, filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"), "# pr reviewer\n")
	writeFile(t, filepath.Join(installedRoot, "commit-helper", "SKILL.md"), "# commit helper\n")

	p := New(t.TempDir(), installedRoot)
	resources, err := p.LocalState()
	if err != nil {
		t.Fatalf("LocalState: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("LocalState() = %+v, want 2 resources", resources)
	}
	byName := map[string]core.Resource{}
	for _, r := range resources {
		byName[r.Name] = r
	}
	if r, ok := byName["pr-reviewer"]; !ok || r.Kind != core.ResourceKindSkill || r.Version == "" {
		t.Fatalf("pr-reviewer resource = %+v, want ResourceKindSkill with non-empty hash", r)
	}
	if byName["pr-reviewer"].SourceMTime == nil {
		t.Fatal("pr-reviewer SourceMTime is nil, want populated")
	}
}

func TestLocalState_EmptyInstalledDirReturnsEmpty(t *testing.T) {
	installedRoot := filepath.Join(t.TempDir(), "does-not-exist")

	p := New(t.TempDir(), installedRoot)
	resources, err := p.LocalState()
	if err != nil {
		t.Fatalf("LocalState: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("LocalState() = %+v, want empty", resources)
	}
}

func TestLocalState_IgnoresNonDirectoryEntries(t *testing.T) {
	installedRoot := t.TempDir()
	writeFile(t, filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"), "# ok\n")
	if err := os.WriteFile(filepath.Join(installedRoot, "stray-file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write stray file: %v", err)
	}

	p := New(t.TempDir(), installedRoot)
	resources, err := p.LocalState()
	if err != nil {
		t.Fatalf("LocalState: %v", err)
	}
	if len(resources) != 1 || resources[0].Name != "pr-reviewer" {
		t.Fatalf("LocalState() = %+v, want only pr-reviewer", resources)
	}
}
