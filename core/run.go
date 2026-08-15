package core

import (
	"fmt"
	"sort"
)

// Run is the generic orchestration entrypoint: it queries every
// provider's local and desired state, computes a per-provider diff,
// concatenates the results into one combined selection screen, and —
// on confirmation — applies the selected items and persists the lock
// file.
//
// When autoApply is true, the interactive select/confirm step is
// skipped entirely and every diff item is applied — for non-TTY
// contexts (e.g. a Docker build) where no one is present to answer
// prompts.
func Run(providers []Provider, lockPath string, autoApply bool) error {
	lock, lockErr := LoadLock(lockPath)
	if lockErr != nil {
		fmt.Printf("warning: could not read lock file (%v); continuing without cache\n", lockErr)
	}

	var allDiffs []DiffItem
	byKind := make(map[ResourceKind]Provider, len(providers))
	for _, p := range providers {
		byKind[p.Kind()] = p

		local, err := p.LocalState()
		if err != nil {
			return fmt.Errorf("%s: local state: %w", p.Kind(), err)
		}
		desired, err := p.DesiredState()
		if err != nil {
			return fmt.Errorf("%s: desired state: %w", p.Kind(), err)
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

	if len(allDiffs) == 0 {
		fmt.Println("Everything is already in sync.")
		return nil
	}

	sortDiffItems(allDiffs)

	var selected []DiffItem
	if autoApply {
		selected = allDiffs
		fmt.Printf("Auto-applying %d change(s) (non-interactive mode).\n", len(selected))
	} else {
		var confirmed bool
		var err error
		selected, confirmed, err = SelectAndConfirm(allDiffs)
		if err != nil {
			return fmt.Errorf("interactive selection: %w", err)
		}
		if !confirmed || len(selected) == 0 {
			fmt.Println("No changes applied.")
			return nil
		}
	}

	results := applyAll(selected, byKind)
	printApplyResults(results)

	updateLock(&lock, results)
	if err := SaveLock(lockPath, lock); err != nil {
		return fmt.Errorf("save lock: %w", err)
	}
	return nil
}

// sortDiffItems sorts the combined diff list by ResourceKind then
// Name, giving a stable, grouped-looking presentation in the single
// flat selection screen without actual UI grouping.
func sortDiffItems(items []DiffItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].ResourceKind != items[j].ResourceKind {
			return items[i].ResourceKind < items[j].ResourceKind
		}
		return items[i].Name < items[j].Name
	})
}

// applyAll applies each selected DiffItem via the provider matching
// its ResourceKind, as a best-effort batch: one item's failure does
// not prevent the remaining items from being attempted, and every
// outcome is reported (never silently skipped).
func applyAll(selected []DiffItem, byKind map[ResourceKind]Provider) []ApplyResult {
	results := make([]ApplyResult, 0, len(selected))
	for _, item := range selected {
		provider, ok := byKind[item.ResourceKind]
		if !ok {
			results = append(results, ApplyResult{
				Item: item,
				Err:  fmt.Errorf("no provider registered for resource kind %q", item.ResourceKind),
			})
			continue
		}
		err := provider.Apply(item)
		results = append(results, ApplyResult{Item: item, Err: err})
	}
	return results
}

func printApplyResults(results []ApplyResult) {
	fmt.Println("\nApply results:")
	failures := 0
	for _, r := range results {
		if r.Err != nil {
			failures++
			fmt.Printf("  FAILED  [%s] %s (%s): %v\n", r.Item.ResourceKind, r.Item.Name, r.Item.Kind, r.Err)
			continue
		}
		fmt.Printf("  OK      [%s] %s (%s)\n", r.Item.ResourceKind, r.Item.Name, r.Item.Kind)
	}
	fmt.Printf("\n%d/%d succeeded, %d failed.\n", len(results)-failures, len(results), failures)
}
