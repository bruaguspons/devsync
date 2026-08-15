package core

import "testing"

func TestResource_ZeroValue(t *testing.T) {
	var r Resource
	if r.Kind != "" || r.Name != "" || r.Version != "" || r.SourceMTime != nil {
		t.Fatalf("zero-value Resource = %+v, want all zero", r)
	}
}

func TestResource_Construction(t *testing.T) {
	r := Resource{Kind: ResourceKindTool, Name: "node", Version: "24.0.0"}
	if r.Kind != ResourceKindTool || r.Name != "node" || r.Version != "24.0.0" {
		t.Fatalf("Resource construction = %+v, unexpected fields", r)
	}
}

func TestDiffItem_ZeroValue(t *testing.T) {
	var d DiffItem
	if d.ResourceKind != "" || d.Name != "" || d.Kind != "" || d.DesiredMTime != nil {
		t.Fatalf("zero-value DiffItem = %+v, want all zero", d)
	}
}

func TestDiffItem_Construction(t *testing.T) {
	d := DiffItem{
		ResourceKind:   ResourceKindSkill,
		Name:           "pr-reviewer",
		LocalVersion:   "aaa",
		DesiredVersion: "bbb",
		Kind:           KindUpdate,
	}
	if d.ResourceKind != ResourceKindSkill || d.Name != "pr-reviewer" ||
		d.LocalVersion != "aaa" || d.DesiredVersion != "bbb" || d.Kind != KindUpdate {
		t.Fatalf("DiffItem construction = %+v, unexpected fields", d)
	}
}

func TestApplyResult_Construction(t *testing.T) {
	item := DiffItem{Name: "x", Kind: KindNew}
	r := ApplyResult{Item: item, Err: nil}
	if r.Item.Name != "x" || r.Err != nil {
		t.Fatalf("ApplyResult construction = %+v, unexpected fields", r)
	}
}
