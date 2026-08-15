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
