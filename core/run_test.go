package core

import (
	"path/filepath"
	"testing"
)

// fakeProvider is an in-package test double implementing Provider,
// used to verify core.Run's orchestration without any real mise/skills
// backend.
type fakeProvider struct {
	kind       ResourceKind
	local      []Resource
	desired    []Resource
	localErr   error
	desiredErr error
	applyErr   map[string]error // by resource name
	applyCalls []DiffItem
}

func (f *fakeProvider) Kind() ResourceKind { return f.kind }

func (f *fakeProvider) LocalState() ([]Resource, error) {
	return f.local, f.localErr
}

func (f *fakeProvider) DesiredState() ([]Resource, error) {
	return f.desired, f.desiredErr
}

func (f *fakeProvider) Apply(item DiffItem) error {
	f.applyCalls = append(f.applyCalls, item)
	if f.applyErr != nil {
		if err, ok := f.applyErr[item.Name]; ok {
			return err
		}
	}
	return nil
}

func TestRun_CombinedDiffAggregatesAcrossProviders(t *testing.T) {
	toolProvider := &fakeProvider{
		kind:    ResourceKindTool,
		local:   []Resource{},
		desired: []Resource{{Kind: ResourceKindTool, Name: "node", Version: "24.0.0"}},
	}
	skillProvider := &fakeProvider{
		kind:    ResourceKindSkill,
		local:   []Resource{},
		desired: []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "abc123"}},
	}

	lockPath := filepath.Join(t.TempDir(), "lock.json")

	// SelectAndConfirm is interactive (huh forms); Run cannot be driven
	// through the interactive selection path in a unit test. This test
	// instead verifies the pre-selection aggregation contract directly
	// by replicating Run's diff-gathering loop, which is the part under
	// test per the "combined diff aggregation" requirement.
	lock, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}

	var allDiffs []DiffItem
	for _, p := range []Provider{toolProvider, skillProvider} {
		local, err := p.LocalState()
		if err != nil {
			t.Fatalf("LocalState: %v", err)
		}
		desired, err := p.DesiredState()
		if err != nil {
			t.Fatalf("DesiredState: %v", err)
		}
		local = reconcileWithLock(local, lock)
		allDiffs = append(allDiffs, Diff(local, desired)...)
	}

	if len(allDiffs) != 2 {
		t.Fatalf("combined diffs = %+v, want 2 items (one tool, one skill)", allDiffs)
	}

	kinds := map[ResourceKind]bool{}
	for _, d := range allDiffs {
		kinds[d.ResourceKind] = true
	}
	if !kinds[ResourceKindTool] || !kinds[ResourceKindSkill] {
		t.Fatalf("combined diffs = %+v, want both tool and skill kinds present", allDiffs)
	}
}

func TestRun_OwnershipPartitioning(t *testing.T) {
	skillProvider := &fakeProvider{
		kind: ResourceKindSkill,
		local: []Resource{
			{Kind: ResourceKindSkill, Name: "orphan-skill", Version: "hash-orphan"},
			{Kind: ResourceKindSkill, Name: "adopted-skill", Version: "matching-hash"},
			{Kind: ResourceKindSkill, Name: "conflicting-skill", Version: "local-hash"},
		},
		desired: []Resource{
			{Kind: ResourceKindSkill, Name: "adopted-skill", Version: "matching-hash"},
			{Kind: ResourceKindSkill, Name: "conflicting-skill", Version: "desired-hash"},
		},
	}

	lockPath := filepath.Join(t.TempDir(), "lock.json")

	// Run() cannot be driven through its interactive SelectAndConfirm
	// path in a unit test; replicate Run's diff-gathering loop with the
	// ownership partitioning step wired in, matching the intended
	// core/run.go change.
	lock, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}

	var allDiffs []DiffItem
	for _, p := range []Provider{skillProvider} {
		local, err := p.LocalState()
		if err != nil {
			t.Fatalf("LocalState: %v", err)
		}
		desired, err := p.DesiredState()
		if err != nil {
			t.Fatalf("DesiredState: %v", err)
		}

		local = reconcileWithLock(local, lock)
		owned, conflicts, adopted := partitionOwnership(local, desired, lock)
		lock.Resources = append(lock.Resources, adopted...)

		diffs := Diff(owned, desired)
		for i := range diffs {
			if conflicts[diffs[i].Name] {
				diffs[i].OwnershipWarning = true
			}
		}
		allDiffs = append(allDiffs, diffs...)
	}

	if err := SaveLock(lockPath, lock); err != nil {
		t.Fatalf("SaveLock: %v", err)
	}

	// (a) foreign local resource absent from desired never appears in allDiffs.
	for _, d := range allDiffs {
		if d.Name == "orphan-skill" {
			t.Fatalf("allDiffs = %+v, orphan-skill (foreign, no desired match) must not appear", allDiffs)
		}
	}

	// (b) foreign local resource whose Version exactly matches desired
	// produces zero diff items and appears in lock.Resources.
	for _, d := range allDiffs {
		if d.Name == "adopted-skill" {
			t.Fatalf("allDiffs = %+v, adopted-skill (exact match) must produce zero diff items", allDiffs)
		}
	}
	reloaded, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock (reload): %v", err)
	}
	foundAdopted := false
	for _, item := range reloaded.Resources {
		if item.Kind == ResourceKindSkill && item.Name == "adopted-skill" {
			foundAdopted = true
		}
	}
	if !foundAdopted {
		t.Fatalf("reloaded lock resources = %+v, want adopted-skill registered", reloaded.Resources)
	}

	// (c) foreign local resource sharing a desired name but differing
	// content produces exactly one KindNew DiffItem with OwnershipWarning: true.
	var conflictItem *DiffItem
	for i := range allDiffs {
		if allDiffs[i].Name == "conflicting-skill" {
			conflictItem = &allDiffs[i]
		}
	}
	if conflictItem == nil {
		t.Fatalf("allDiffs = %+v, want a diff item for conflicting-skill", allDiffs)
	}
	if conflictItem.Kind != KindNew {
		t.Fatalf("conflicting-skill diff kind = %v, want KindNew", conflictItem.Kind)
	}
	if !conflictItem.OwnershipWarning {
		t.Fatalf("conflicting-skill diff = %+v, want OwnershipWarning: true", conflictItem)
	}
}

func TestApplyAll_RoutesByResourceKind(t *testing.T) {
	toolProvider := &fakeProvider{kind: ResourceKindTool}
	skillProvider := &fakeProvider{kind: ResourceKindSkill}
	byKind := map[ResourceKind]Provider{
		ResourceKindTool:  toolProvider,
		ResourceKindSkill: skillProvider,
	}

	selected := []DiffItem{
		{ResourceKind: ResourceKindTool, Name: "node", Kind: KindNew},
		{ResourceKind: ResourceKindSkill, Name: "pr-reviewer", Kind: KindNew},
	}

	results := applyAll(selected, byKind)

	if len(results) != 2 {
		t.Fatalf("applyAll() = %+v, want 2 results", results)
	}
	if len(toolProvider.applyCalls) != 1 || toolProvider.applyCalls[0].Name != "node" {
		t.Fatalf("tool provider applyCalls = %+v, want exactly [node]", toolProvider.applyCalls)
	}
	if len(skillProvider.applyCalls) != 1 || skillProvider.applyCalls[0].Name != "pr-reviewer" {
		t.Fatalf("skill provider applyCalls = %+v, want exactly [pr-reviewer]", skillProvider.applyCalls)
	}
}

func TestApplyAll_UnknownResourceKindReportsError(t *testing.T) {
	byKind := map[ResourceKind]Provider{}
	selected := []DiffItem{{ResourceKind: ResourceKindTool, Name: "node", Kind: KindNew}}

	results := applyAll(selected, byKind)

	if len(results) != 1 || results[0].Err == nil {
		t.Fatalf("applyAll() = %+v, want one errored result for unrouted kind", results)
	}
}

func TestUpdateLock_AfterApply_RecordsOnlySuccesses(t *testing.T) {
	lock := LockFile{}
	results := []ApplyResult{
		{Item: DiffItem{ResourceKind: ResourceKindTool, Name: "node", DesiredVersion: "24.0.0", Kind: KindNew}, Err: nil},
		{Item: DiffItem{ResourceKind: ResourceKindSkill, Name: "broken-skill", DesiredVersion: "xyz", Kind: KindNew}, Err: errBoom},
	}

	updateLock(&lock, results)

	if len(lock.Resources) != 1 || lock.Resources[0].Name != "node" {
		t.Fatalf("updateLock() after apply = %+v, want only the successful node entry", lock.Resources)
	}
}

var errBoom = simpleErr("boom")

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
