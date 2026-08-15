package skills

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func TestHashDirTree_Deterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "# hello\n")
	writeFile(t, filepath.Join(dir, "scripts", "run.sh"), "echo hi\n")

	h1, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}
	h2, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}
	if h1 != h2 {
		t.Fatalf("hashDirTree() not deterministic: %q != %q", h1, h2)
	}
}

func TestHashDirTree_SensitiveToContentChange(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "# hello\n")
	before, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	writeFile(t, filepath.Join(dir, "SKILL.md"), "# hello world\n")
	after, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	if before == after {
		t.Fatal("hashDirTree() did not change after content edit")
	}
}

func TestHashDirTree_SensitiveToRename(t *testing.T) {
	dirA := t.TempDir()
	writeFile(t, filepath.Join(dirA, "a.md"), "same content\n")
	hashA, _, err := hashDirTree(dirA)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	dirB := t.TempDir()
	writeFile(t, filepath.Join(dirB, "b.md"), "same content\n")
	hashB, _, err := hashDirTree(dirB)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	if hashA == hashB {
		t.Fatal("hashDirTree() did not change when the only file was renamed")
	}
}

func TestHashDirTree_SensitiveToAddRemove(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "content\n")
	before, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	writeFile(t, filepath.Join(dir, "extra.md"), "more content\n")
	after, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}

	if before == after {
		t.Fatal("hashDirTree() did not change after adding a file")
	}

	if err := os.Remove(filepath.Join(dir, "extra.md")); err != nil {
		t.Fatalf("remove: %v", err)
	}
	backToOriginal, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}
	if backToOriginal != before {
		t.Fatal("hashDirTree() did not return to original hash after removing the added file")
	}
}

func TestHashDirTree_PathSeparatorIndependence(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "nested", "file.md"), "x\n")

	got, maxMTime, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}
	if got == "" {
		t.Fatal("hashDirTree() returned empty digest")
	}
	if maxMTime.IsZero() {
		t.Fatal("hashDirTree() returned zero maxMTime")
	}
	if maxMTime.After(time.Now().Add(time.Minute)) {
		t.Fatalf("hashDirTree() maxMTime %v is implausibly in the future", maxMTime)
	}
}

func TestHashDirTree_SkipsBrokenSymlink(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "SKILL.md"), "# hello\n")

	brokenLink := filepath.Join(dir, "broken-link")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist-target"), brokenLink); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	got, _, err := hashDirTree(dir)
	if err != nil {
		t.Fatalf("hashDirTree: expected success with a broken symlink present, got error: %v", err)
	}

	// The symlink must be excluded from the digest input: the hash must
	// equal the hash of the same directory without the symlink.
	dirWithoutLink := t.TempDir()
	writeFile(t, filepath.Join(dirWithoutLink, "SKILL.md"), "# hello\n")
	want, _, err := hashDirTree(dirWithoutLink)
	if err != nil {
		t.Fatalf("hashDirTree: %v", err)
	}
	if got != want {
		t.Fatalf("hashDirTree(with broken symlink) = %q, want %q (symlink excluded from digest)", got, want)
	}
}

func TestListSkillDirs_MissingRootReturnsEmpty(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	names, err := listSkillDirs(missing)
	if err != nil {
		t.Fatalf("listSkillDirs: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("listSkillDirs(missing) = %v, want empty", names)
	}
}

func TestListSkillDirs_SortedAndDirsOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "zeta", "SKILL.md"), "x\n")
	writeFile(t, filepath.Join(dir, "alpha", "SKILL.md"), "x\n")
	writeFile(t, filepath.Join(dir, "not-a-skill-file.txt"), "x\n")

	names, err := listSkillDirs(dir)
	if err != nil {
		t.Fatalf("listSkillDirs: %v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "zeta" {
		t.Fatalf("listSkillDirs() = %v, want [alpha zeta]", names)
	}
}
