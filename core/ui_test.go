package core

import (
	"strings"
	"testing"
)

func TestDiffItemLabel_OwnershipWarning(t *testing.T) {
	tests := []struct {
		name             string
		item             DiffItem
		wantSuffixWarn   bool
		wantUnchangedFor DiffItem
	}{
		{
			name: "KindNew with OwnershipWarning true renders warning suffix",
			item: DiffItem{
				ResourceKind:     ResourceKindSkill,
				Name:             "pr-reviewer",
				DesiredVersion:   "abc123",
				Kind:             KindNew,
				OwnershipWarning: true,
			},
			wantSuffixWarn: true,
		},
		{
			name: "KindNew with OwnershipWarning false renders unchanged",
			item: DiffItem{
				ResourceKind:     ResourceKindSkill,
				Name:             "pr-reviewer",
				DesiredVersion:   "abc123",
				Kind:             KindNew,
				OwnershipWarning: false,
			},
			wantSuffixWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffItemLabel(tt.item)

			baseline := tt.item
			baseline.OwnershipWarning = false
			baselineLabel := diffItemLabel(baseline)

			if tt.wantSuffixWarn {
				if got == baselineLabel {
					t.Fatalf("diffItemLabel(%+v) = %q, want a warning suffix distinguishing it from the unwarned label %q", tt.item, got, baselineLabel)
				}
				if !containsUnmanagedWarning(got) {
					t.Fatalf("diffItemLabel(%+v) = %q, want it to contain an ownership warning about unmanaged content", tt.item, got)
				}
			} else {
				if got != baselineLabel {
					t.Fatalf("diffItemLabel(%+v) = %q, want byte-identical to current unwarned label %q", tt.item, got, baselineLabel)
				}
			}
		})
	}
}

func containsUnmanagedWarning(s string) bool {
	needles := []string{"unmanaged", "not managed", "overwrite"}
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}
