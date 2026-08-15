package core

import (
	"sort"
	"testing"
)

func sortedItems(items []DiffItem) []DiffItem {
	sorted := make([]DiffItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return sorted
}

func tool(name, version string) Resource {
	return Resource{Kind: ResourceKindTool, Name: name, Version: version}
}

func TestDiff_NewInRemote(t *testing.T) {
	local := []Resource{}
	desired := []Resource{tool("node", "24.0.0")}

	got := Diff(local, desired)
	if len(got) != 1 {
		t.Fatalf("Diff() = %+v, want 1 item", got)
	}
	want := DiffItem{ResourceKind: ResourceKindTool, Name: "node", DesiredVersion: "24.0.0", Kind: KindNew}
	if got[0].ResourceKind != want.ResourceKind || got[0].Name != want.Name ||
		got[0].DesiredVersion != want.DesiredVersion || got[0].Kind != want.Kind {
		t.Fatalf("Diff() = %+v, want %+v", got[0], want)
	}
}

func TestDiff_RemovedFromRemote(t *testing.T) {
	local := []Resource{tool("watchexec", "1.0.0")}
	desired := []Resource{}

	got := Diff(local, desired)
	if len(got) != 1 {
		t.Fatalf("Diff() = %+v, want 1 item", got)
	}
	want := DiffItem{ResourceKind: ResourceKindTool, Name: "watchexec", LocalVersion: "1.0.0", Kind: KindRemoved}
	if got[0].ResourceKind != want.ResourceKind || got[0].Name != want.Name ||
		got[0].LocalVersion != want.LocalVersion || got[0].Kind != want.Kind {
		t.Fatalf("Diff() = %+v, want %+v", got[0], want)
	}
}

func TestDiff_VersionDrift(t *testing.T) {
	local := []Resource{tool("rust", "1.96.0")}
	desired := []Resource{tool("rust", "1.97.1")}

	got := Diff(local, desired)
	if len(got) != 1 {
		t.Fatalf("Diff() = %+v, want 1 item", got)
	}
	want := DiffItem{ResourceKind: ResourceKindTool, Name: "rust", LocalVersion: "1.96.0", DesiredVersion: "1.97.1", Kind: KindUpdate}
	if got[0].ResourceKind != want.ResourceKind || got[0].Name != want.Name ||
		got[0].LocalVersion != want.LocalVersion || got[0].DesiredVersion != want.DesiredVersion || got[0].Kind != want.Kind {
		t.Fatalf("Diff() = %+v, want %+v", got[0], want)
	}
}

func TestDiff_NoOpNotIncluded(t *testing.T) {
	local := []Resource{tool("node", "24.0.0")}
	desired := []Resource{tool("node", "24.0.0")}

	got := Diff(local, desired)
	if len(got) != 0 {
		t.Fatalf("Diff() = %+v, want empty (no-op)", got)
	}
}

func TestDiff_MixedCategoriesAllPresent(t *testing.T) {
	local := []Resource{
		tool("watchexec", "1.0.0"), // removed
		tool("rust", "1.96.0"),     // update
		tool("node", "24.0.0"),     // no-op
	}
	desired := []Resource{
		tool("rust", "1.97.1"), // update
		tool("node", "24.0.0"), // no-op
		tool("bun", "1.3.0"),   // new
	}

	got := sortedItems(Diff(local, desired))
	if len(got) != 3 {
		t.Fatalf("expected 3 actionable items, got %d: %+v", len(got), got)
	}
	if got[0].Name != "bun" || got[0].Kind != KindNew {
		t.Errorf("got[0] = %+v, want bun/new", got[0])
	}
	if got[1].Name != "rust" || got[1].Kind != KindUpdate {
		t.Errorf("got[1] = %+v, want rust/update", got[1])
	}
	if got[2].Name != "watchexec" || got[2].Kind != KindRemoved {
		t.Errorf("got[2] = %+v, want watchexec/removed", got[2])
	}
}

func TestDiff_ComparisonVersion_AliasMatchIsNoOp(t *testing.T) {
	local := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "24.19.0", ComparisonVersion: "lts"}}
	desired := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "lts", ComparisonVersion: "lts"}}

	got := Diff(local, desired)
	if len(got) != 0 {
		t.Fatalf("Diff() = %+v, want empty (ComparisonVersion match is a no-op)", got)
	}
}

func TestDiff_ComparisonVersion_MismatchIsUpdate(t *testing.T) {
	local := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "22.0.0", ComparisonVersion: "22"}}
	desired := []Resource{{Kind: ResourceKindTool, Name: "node", Version: "lts", ComparisonVersion: "lts"}}

	got := Diff(local, desired)
	if len(got) != 1 || got[0].Kind != KindUpdate {
		t.Fatalf("Diff() = %+v, want 1 KindUpdate item", got)
	}
}

func TestDiff_ComparisonVersion_OneSideEmptyFallsBackToVersion(t *testing.T) {
	// Skills path: ComparisonVersion is never set, must behave exactly
	// as before (plain Version comparison).
	local := []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "abc123"}}
	desired := []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "abc123"}}

	got := Diff(local, desired)
	if len(got) != 0 {
		t.Fatalf("Diff() = %+v, want empty (fallback to Version when ComparisonVersion unset)", got)
	}

	desiredChanged := []Resource{{Kind: ResourceKindSkill, Name: "pr-reviewer", Version: "def456"}}
	got2 := Diff(local, desiredChanged)
	if len(got2) != 1 || got2[0].Kind != KindUpdate {
		t.Fatalf("Diff() = %+v, want 1 KindUpdate item (fallback to Version)", got2)
	}
}

// TestDiff_ResourceKindOblivious asserts Diff() itself does not reject
// or special-case a mixed-kind input slice — the invariant that a
// single Diff() call only ever receives one ResourceKind lives at the
// call site (core.Run), not inside Diff.
func TestDiff_ResourceKindOblivious(t *testing.T) {
	local := []Resource{
		{Kind: ResourceKindTool, Name: "rust", Version: "1.96.0"},
	}
	desired := []Resource{
		{Kind: ResourceKindSkill, Name: "rust", Version: "1.97.1"},
	}

	got := Diff(local, desired)
	if len(got) != 1 || got[0].Kind != KindUpdate {
		t.Fatalf("Diff() with mixed-kind input = %+v, want 1 update item (Diff does not branch on ResourceKind)", got)
	}
}
