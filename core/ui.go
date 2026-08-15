package core

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Per-kind and ownership-warning badge colors for the diff checklist,
// following a traffic-light palette adapted to light/dark terminals.
var (
	styleKindNew       = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0A7A2A", Dark: "#3ECF6E"})
	styleKindUpdate    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8A5A00", Dark: "#E5C300"})
	styleKindRemoved   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B00020", Dark: "#FF6B6B"})
	styleOwnershipWarn = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#B35A00", Dark: "#FFA94D"})
)

// styleForKind returns the badge style for a DiffKind. Unknown kinds get
// the zero-value style (no coloring).
func styleForKind(k DiffKind) lipgloss.Style {
	switch k {
	case KindNew:
		return styleKindNew
	case KindUpdate:
		return styleKindUpdate
	case KindRemoved:
		return styleKindRemoved
	default:
		return lipgloss.Style{}
	}
}

// checklistTheme builds a huh.Theme derived from huh.ThemeBase with
// explicit, visually distinct checkbox prefix glyphs for selected and
// unselected multi-select rows.
func checklistTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.SelectedPrefix = lipgloss.NewStyle().SetString("[x] ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("[ ] ")
	t.Blurred.SelectedPrefix = lipgloss.NewStyle().SetString("[x] ")
	t.Blurred.UnselectedPrefix = lipgloss.NewStyle().SetString("[ ] ")
	return t
}

// resourceKindLabel returns a short bracketed prefix identifying a
// DiffItem's resource kind for display purposes.
func resourceKindLabel(kind ResourceKind) string {
	switch kind {
	case ResourceKindTool:
		return "[tool]"
	case ResourceKindSkill:
		return "[skill]"
	default:
		return "[" + string(kind) + "]"
	}
}

// diffItemLabel renders a human-readable label for one DiffItem in the
// multi-select checklist, prefixed by its ResourceKind.
func diffItemLabel(item DiffItem) string {
	prefix := resourceKindLabel(item.ResourceKind)
	style := styleForKind(item.Kind)
	var label string
	switch item.Kind {
	case KindNew:
		label = fmt.Sprintf("%s%s %s@%s", prefix, style.Render("[install]"), item.Name, item.DesiredVersion)
	case KindRemoved:
		label = fmt.Sprintf("%s%s %s@%s", prefix, style.Render("[uninstall]"), item.Name, item.LocalVersion)
	case KindUpdate:
		label = fmt.Sprintf("%s%s %s: %s -> %s", prefix, style.Render("[update available]"), item.Name, item.LocalVersion, item.DesiredVersion)
	default:
		label = prefix + item.Name
	}
	if item.OwnershipWarning {
		label += styleOwnershipWarn.Render(" (warning: not managed by devsync — apply will overwrite unmanaged content)")
	}
	return label
}

// SelectAndConfirm renders a huh multi-select checklist over the given
// diff items — a single flat list spanning every resource kind, kind
// visible per-row via the label prefix — then an explicit confirm
// step. It returns the subset the user selected only if the user
// explicitly confirmed; otherwise it returns nil, false (no apply must
// occur).
//
// If diffItems is empty, there is nothing to select or confirm.
func SelectAndConfirm(diffItems []DiffItem) (selected []DiffItem, confirmed bool, err error) {
	if len(diffItems) == 0 {
		return nil, false, nil
	}

	options := make([]huh.Option[int], len(diffItems))
	for i, item := range diffItems {
		options[i] = huh.NewOption(diffItemLabel(item), i)
	}

	var chosenIdx []int
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select changes to apply").
				Description("Space to toggle, enter to continue. Nothing is applied without explicit confirmation.").
				Options(options...).
				Value(&chosenIdx),
		),
	).WithTheme(checklistTheme())
	if err := selectForm.Run(); err != nil {
		return nil, false, fmt.Errorf("selection form: %w", err)
	}
	if len(chosenIdx) == 0 {
		return nil, false, nil
	}

	chosen := make([]DiffItem, 0, len(chosenIdx))
	for _, idx := range chosenIdx {
		chosen = append(chosen, diffItems[idx])
	}

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Apply %d selected change(s)?", len(chosen))).
				Affirmative("Yes, apply").
				Negative("No, cancel").
				Value(&confirm),
		),
	)
	if err := confirmForm.Run(); err != nil {
		return nil, false, fmt.Errorf("confirm form: %w", err)
	}
	if !confirm {
		return nil, false, nil
	}

	return chosen, true, nil
}
