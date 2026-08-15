package skills

import (
	"path/filepath"
	"testing"

	"github.com/bruaguspons/devsync/core"
)

func TestDesiredState_WalksDesiredRootSkills(t *testing.T) {
	desiredRoot := t.TempDir()
	writeFile(t, filepath.Join(desiredRoot, "pr-reviewer", "SKILL.md"), "# pr reviewer\n")

	p := New(desiredRoot, t.TempDir())
	resources, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}

	if len(resources) != 1 || resources[0].Name != "pr-reviewer" || resources[0].Kind != core.ResourceKindSkill {
		t.Fatalf("DesiredState() = %+v, want one pr-reviewer skill resource", resources)
	}
	if resources[0].Version == "" {
		t.Fatal("DesiredState() returned empty hash")
	}
}

func TestDesiredState_EmptyDesiredRootReturnsEmpty(t *testing.T) {
	desiredRoot := filepath.Join(t.TempDir(), "does-not-exist")

	p := New(desiredRoot, t.TempDir())
	resources, err := p.DesiredState()
	if err != nil {
		t.Fatalf("DesiredState: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("DesiredState() = %+v, want empty", resources)
	}
}
