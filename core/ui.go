package core

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

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
	switch item.Kind {
	case KindNew:
		return fmt.Sprintf("%s[install] %s@%s", prefix, item.Name, item.DesiredVersion)
	case KindRemoved:
		return fmt.Sprintf("%s[uninstall] %s@%s", prefix, item.Name, item.LocalVersion)
	case KindUpdate:
		return fmt.Sprintf("%s[update available] %s: %s -> %s", prefix, item.Name, item.LocalVersion, item.DesiredVersion)
	default:
		return prefix + item.Name
	}
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
	)
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
