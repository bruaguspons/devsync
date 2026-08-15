package skills

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bruaguspons/devsync/core"
)

func TestApply_Install_CopiesAndPreservesSubdirectories(t *testing.T) {
	desiredRoot := t.TempDir()
	installedRoot := t.TempDir()
	writeFile(t, filepath.Join(desiredRoot, "pr-reviewer", "SKILL.md"), "# pr reviewer\n")
	writeFile(t, filepath.Join(desiredRoot, "pr-reviewer", "scripts", "run.sh"), "echo hi\n")

	p := New(desiredRoot, installedRoot)
	err := p.Apply(core.DiffItem{ResourceKind: core.ResourceKindSkill, Name: "pr-reviewer", Kind: core.KindNew})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, err := os.Stat(filepath.Join(installedRoot, "pr-reviewer", "SKILL.md")); err != nil {
		t.Fatalf("expected SKILL.md installed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(installedRoot, "pr-reviewer", "scripts", "run.sh")); err != nil {
		t.Fatalf("expected nested scripts/run.sh installed: %v", err)
	}
}

func TestApply_Update_RemovesStaleExtraDestFile(t *testing.T) {
	desiredRoot := t.TempDir()
	installedRoot := t.TempDir()
	writeFile(t, filepath.Join(desiredRoot, "pr-reviewer", "SKILL.md"), "# updated\n")
	// Pre-existing install has an extra file no longer in the repo source.
	writeFile(t, filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"), "# old\n")
	writeFile(t, filepath.Join(installedRoot, "pr-reviewer", "stale.md"), "leftover\n")

	p := New(desiredRoot, installedRoot)
	err := p.Apply(core.DiffItem{ResourceKind: core.ResourceKindSkill, Name: "pr-reviewer", Kind: core.KindUpdate})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, err := os.Stat(filepath.Join(installedRoot, "pr-reviewer", "stale.md")); !os.IsNotExist(err) {
		t.Fatalf("expected stale.md removed by RemoveAll-then-copy update strategy, stat err = %v", err)
	}
	content, err := os.ReadFile(filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatalf("read updated SKILL.md: %v", err)
	}
	if string(content) != "# updated\n" {
		t.Fatalf("SKILL.md content = %q, want updated content", content)
	}
}

func TestApply_Install_PreservesSourceMTime(t *testing.T) {
	desiredRoot := t.TempDir()
	installedRoot := t.TempDir()
	srcPath := filepath.Join(desiredRoot, "pr-reviewer", "SKILL.md")
	writeFile(t, srcPath, "# pr reviewer\n")

	sourceMTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(srcPath, sourceMTime, sourceMTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	p := New(desiredRoot, installedRoot)
	if err := p.Apply(core.DiffItem{ResourceKind: core.ResourceKindSkill, Name: "pr-reviewer", Kind: core.KindNew}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	dstInfo, err := os.Stat(filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"))
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}
	if !dstInfo.ModTime().Equal(sourceMTime) {
		t.Fatalf("dest mtime = %v, want source mtime %v", dstInfo.ModTime(), sourceMTime)
	}
}

func TestApply_Remove_DeletesOnlyNamedSkillDirectory(t *testing.T) {
	installedRoot := t.TempDir()
	writeFile(t, filepath.Join(installedRoot, "pr-reviewer", "SKILL.md"), "# gone soon\n")
	writeFile(t, filepath.Join(installedRoot, "other-skill", "SKILL.md"), "# keep me\n")

	p := New(t.TempDir(), installedRoot)
	err := p.Apply(core.DiffItem{ResourceKind: core.ResourceKindSkill, Name: "pr-reviewer", Kind: core.KindRemoved})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, err := os.Stat(filepath.Join(installedRoot, "pr-reviewer")); !os.IsNotExist(err) {
		t.Fatalf("expected pr-reviewer directory removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(installedRoot, "other-skill", "SKILL.md")); err != nil {
		t.Fatalf("expected other-skill left untouched: %v", err)
	}
}
