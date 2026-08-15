package core

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// wantPlainLabel independently reconstructs the pre-colorization plain-text
// label content, so TestDiffItemLabel_ANSIStyled proves styling is additive
// only, without depending on diffItemLabel's own formatting to grade itself.
func wantPlainLabel(item DiffItem) string {
	prefix := resourceKindLabel(item.ResourceKind)
	switch item.Kind {
	case KindNew:
		return prefix + "[install] " + item.Name + "@" + item.DesiredVersion
	case KindRemoved:
		return prefix + "[uninstall] " + item.Name + "@" + item.LocalVersion
	case KindUpdate:
		return prefix + "[update available] " + item.Name + ": " + item.LocalVersion + " -> " + item.DesiredVersion
	default:
		return prefix + item.Name
	}
}

func TestStyleForKind_DistinctForeground(t *testing.T) {
	newFg := styleForKind(KindNew).GetForeground()
	updateFg := styleForKind(KindUpdate).GetForeground()
	removedFg := styleForKind(KindRemoved).GetForeground()

	if newFg == updateFg {
		t.Fatalf("styleForKind(KindNew).GetForeground() == styleForKind(KindUpdate).GetForeground() (%v), want distinct colors", newFg)
	}
	if newFg == removedFg {
		t.Fatalf("styleForKind(KindNew).GetForeground() == styleForKind(KindRemoved).GetForeground() (%v), want distinct colors", newFg)
	}
	if updateFg == removedFg {
		t.Fatalf("styleForKind(KindUpdate).GetForeground() == styleForKind(KindRemoved).GetForeground() (%v), want distinct colors", updateFg)
	}
}

func TestDiffItemLabel_ANSIStyled(t *testing.T) {
	prevProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(prevProfile)

	tests := []struct {
		name string
		item DiffItem
	}{
		{
			name: "KindNew install marker styled",
			item: DiffItem{ResourceKind: ResourceKindSkill, Name: "pr-reviewer", DesiredVersion: "abc123", Kind: KindNew},
		},
		{
			name: "KindUpdate marker styled",
			item: DiffItem{ResourceKind: ResourceKindTool, Name: "go", LocalVersion: "1.21", DesiredVersion: "1.22", Kind: KindUpdate},
		},
		{
			name: "KindRemoved marker styled",
			item: DiffItem{ResourceKind: ResourceKindTool, Name: "node", LocalVersion: "18", Kind: KindRemoved},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffItemLabel(tt.item)

			if !ansiPattern.MatchString(got) {
				t.Fatalf("diffItemLabel(%+v) = %q, want it to contain ANSI styling escape sequences when color profile is TrueColor", tt.item, got)
			}

			stripped := stripANSI(got)
			wantPlain := wantPlainLabel(tt.item)
			if stripped != wantPlain {
				t.Fatalf("stripANSI(diffItemLabel(%+v)) = %q, want %q (plain-text content must be unchanged by styling)", tt.item, stripped, wantPlain)
			}
		})
	}
}

func TestChecklistTheme_PrefixGlyphs(t *testing.T) {
	theme := checklistTheme()

	focusedSelected := theme.Focused.SelectedPrefix.String()
	focusedUnselected := theme.Focused.UnselectedPrefix.String()
	blurredSelected := theme.Blurred.SelectedPrefix.String()
	blurredUnselected := theme.Blurred.UnselectedPrefix.String()

	if focusedSelected == "" {
		t.Fatal("checklistTheme().Focused.SelectedPrefix is empty, want a non-empty glyph")
	}
	if focusedUnselected == "" {
		t.Fatal("checklistTheme().Focused.UnselectedPrefix is empty, want a non-empty glyph")
	}
	if focusedSelected == focusedUnselected {
		t.Fatalf("checklistTheme().Focused.SelectedPrefix (%q) == UnselectedPrefix (%q), want visibly distinct glyphs", focusedSelected, focusedUnselected)
	}
	if blurredSelected == "" || blurredUnselected == "" {
		t.Fatal("checklistTheme().Blurred prefixes must be non-empty")
	}
	if blurredSelected == blurredUnselected {
		t.Fatalf("checklistTheme().Blurred.SelectedPrefix (%q) == UnselectedPrefix (%q), want visibly distinct glyphs", blurredSelected, blurredUnselected)
	}
}

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
