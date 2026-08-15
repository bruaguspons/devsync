package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSaveLock_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock.json")

	now := time.Now().UTC().Truncate(time.Second)
	original := LockFile{
		Version: 1,
		Resources: []LockedItem{
			{Kind: ResourceKindTool, Name: "node", Version: "24.0.0", InstalledAt: now},
			{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "abc123", InstalledAt: now, SourceMTime: &now},
		},
	}

	if err := SaveLock(path, original); err != nil {
		t.Fatalf("SaveLock: %v", err)
	}

	loaded, err := LoadLock(path)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if len(loaded.Resources) != 2 {
		t.Fatalf("LoadLock() = %+v, want 2 resources", loaded)
	}
	if loaded.Resources[0].Name != "node" || loaded.Resources[1].Name != "pr-reviewer" {
		t.Fatalf("LoadLock() resources = %+v, unexpected names", loaded.Resources)
	}
	if loaded.Resources[1].SourceMTime == nil || !loaded.Resources[1].SourceMTime.Equal(now) {
		t.Fatalf("LoadLock() SourceMTime = %v, want %v", loaded.Resources[1].SourceMTime, now)
	}
}

func TestLoadLock_MissingFileReturnsEmpty(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.json")

	lock, err := LoadLock(missing)
	if err != nil {
		t.Fatalf("LoadLock: expected no error for missing file, got %v", err)
	}
	if len(lock.Resources) != 0 {
		t.Fatalf("LoadLock() = %+v, want empty resources", lock)
	}
}

func TestLoadLock_CorruptJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write corrupt lock: %v", err)
	}

	lock, err := LoadLock(path)
	if err == nil {
		t.Fatal("LoadLock: expected error for corrupt JSON, got nil")
	}
	// Non-fatal: callers get an empty, usable LockFile alongside the error.
	if len(lock.Resources) != 0 {
		t.Fatalf("LoadLock() on error = %+v, want empty resources", lock)
	}
}

func TestReconcileWithLock_ToolsAreNoOp(t *testing.T) {
	local := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "24.0.0"}}
	lock := LockFile{Resources: []LockedItem{
		{Kind: ResourceKindTool, Name: "node", Version: "20.0.0"},
	}}

	got := reconcileWithLock(local, lock)
	if len(got) != 1 || got[0].Version != "24.0.0" {
		t.Fatalf("reconcileWithLock(tool) = %+v, want unchanged local (24.0.0)", got)
	}
}

func TestReconcileWithLock_SkillMTimeMatchReusesCachedVersion(t *testing.T) {
	mtime := time.Now().UTC().Truncate(time.Second)
	local := []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "freshly-computed", SourceMTime: &mtime}}
	lock := LockFile{Resources: []LockedItem{
		{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "cached-hash", SourceMTime: &mtime},
	}}

	got := reconcileWithLock(local, lock)
	if len(got) != 1 || got[0].Version != "cached-hash" {
		t.Fatalf("reconcileWithLock(skill, mtime match) = %+v, want cached-hash reused", got)
	}
}

func TestReconcileWithLock_SkillMTimeMismatchKeepsFreshValue(t *testing.T) {
	oldMTime := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	newMTime := time.Now().UTC().Truncate(time.Second)
	local := []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "freshly-computed", SourceMTime: &newMTime}}
	lock := LockFile{Resources: []LockedItem{
		{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "stale-hash", SourceMTime: &oldMTime},
	}}

	got := reconcileWithLock(local, lock)
	if len(got) != 1 || got[0].Version != "freshly-computed" {
		t.Fatalf("reconcileWithLock(skill, mtime mismatch) = %+v, want freshly-computed kept", got)
	}
}

func TestReconcileWithLock_SkillNoLockEntryKeepsFreshValue(t *testing.T) {
	mtime := time.Now().UTC()
	local := []Resource{{Kind: ResourceKindSkill, Name: "brand-new-skill", Version: "freshly-computed", SourceMTime: &mtime}}
	lock := LockFile{}

	got := reconcileWithLock(local, lock)
	if len(got) != 1 || got[0].Version != "freshly-computed" {
		t.Fatalf("reconcileWithLock(skill, no lock entry) = %+v, want freshly-computed kept", got)
	}
}

func TestReconcileWithLock_NilSourceMTimeIsNoOpRegardlessOfKind(t *testing.T) {
	// A tool resource (SourceMTime always nil) must be left unchanged
	// purely because SourceMTime is nil — reconcileWithLock no longer
	// branches on ResourceKind at all, so this proves the mtime guard
	// alone (not a kind check) drives the no-op behavior.
	local := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "24.0.0"}}
	lock := LockFile{Resources: []LockedItem{
		{Kind: ResourceKindTool, Name: "node", Version: "20.0.0"},
	}}

	got := reconcileWithLock(local, lock)
	if len(got) != 1 || got[0].Version != "24.0.0" {
		t.Fatalf("reconcileWithLock(nil SourceMTime) = %+v, want unchanged local (24.0.0)", got)
	}
}

func TestUpdateLock_UpsertsSuccessfulApplies(t *testing.T) {
	lock := LockFile{}
	results := []ApplyResult{
		{Item: DiffItem{ResourceKind: ResourceKindTool, Name: "node", DesiredVersion: "24.0.0", Kind: KindNew}, Err: nil},
	}

	updateLock(&lock, results)

	if len(lock.Resources) != 1 || lock.Resources[0].Name != "node" || lock.Resources[0].Version != "24.0.0" {
		t.Fatalf("updateLock() = %+v, want one upserted node@24.0.0 entry", lock.Resources)
	}
}

func TestUpdateLock_SkipsFailedApplies(t *testing.T) {
	lock := LockFile{}
	results := []ApplyResult{
		{Item: DiffItem{ResourceKind: ResourceKindTool, Name: "node", DesiredVersion: "24.0.0", Kind: KindNew}, Err: os.ErrInvalid},
	}

	updateLock(&lock, results)

	if len(lock.Resources) != 0 {
		t.Fatalf("updateLock() = %+v, want no entries recorded for failed apply", lock.Resources)
	}
}

func TestUpdateLock_RemovedSuccessDeletesEntry(t *testing.T) {
	lock := LockFile{Resources: []LockedItem{
		{Kind: ResourceKindTool, Name: "watchexec", Version: "1.0.0"},
	}}
	results := []ApplyResult{
		{Item: DiffItem{ResourceKind: ResourceKindTool, Name: "watchexec", Kind: KindRemoved}, Err: nil},
	}

	updateLock(&lock, results)

	if len(lock.Resources) != 0 {
		t.Fatalf("updateLock() = %+v, want removed entry deleted", lock.Resources)
	}
}
