package core

import "time"

// Diff computes the set of actionable differences between local
// (installed) resource state and desired (declared) resource state.
// It is resource-kind-oblivious: it never compares a tool against a
// skill because it is only ever called once per provider — all
// Resources passed to a single Diff() call share one ResourceKind.
// That invariant lives at the call site (core.Run), not here.
//
// Classification:
//   - present only in desired  -> KindNew (installable)
//   - present only in local    -> KindRemoved (uninstallable)
//   - present in both, version differs -> KindUpdate (version-drift)
//   - present in both, same version -> not included (no-op, not actionable)
func Diff(local, desired []Resource) []DiffItem {
	localByName := make(map[string]Resource, len(local))
	for _, s := range local {
		localByName[s.Name] = s
	}
	desiredByName := make(map[string]Resource, len(desired))
	desiredMTimeByName := make(map[string]*time.Time, len(desired))
	for _, s := range desired {
		desiredByName[s.Name] = s
		desiredMTimeByName[s.Name] = s.SourceMTime
	}

	var kind ResourceKind
	if len(local) > 0 {
		kind = local[0].Kind
	} else if len(desired) > 0 {
		kind = desired[0].Kind
	}

	seen := make(map[string]bool, len(localByName)+len(desiredByName))
	names := make([]string, 0, len(localByName)+len(desiredByName))
	for _, s := range local {
		if !seen[s.Name] {
			seen[s.Name] = true
			names = append(names, s.Name)
		}
	}
	for _, s := range desired {
		if !seen[s.Name] {
			seen[s.Name] = true
			names = append(names, s.Name)
		}
	}

	items := make([]DiffItem, 0, len(names))
	for _, name := range names {
		localRes, inLocal := localByName[name]
		desiredRes, inDesired := desiredByName[name]
		localVersion, desiredVersion := localRes.Version, desiredRes.Version

		compareLocal, compareDesired := localVersion, desiredVersion
		if inLocal && inDesired && localRes.ComparisonVersion != "" && desiredRes.ComparisonVersion != "" {
			compareLocal, compareDesired = localRes.ComparisonVersion, desiredRes.ComparisonVersion
		}

		switch {
		case inDesired && !inLocal:
			items = append(items, DiffItem{
				ResourceKind:   kind,
				Name:           name,
				DesiredVersion: desiredVersion,
				DesiredMTime:   desiredMTimeByName[name],
				Kind:           KindNew,
			})
		case inLocal && !inDesired:
			items = append(items, DiffItem{
				ResourceKind: kind,
				Name:         name,
				LocalVersion: localVersion,
				Kind:         KindRemoved,
			})
		case inLocal && inDesired && compareLocal != compareDesired:
			items = append(items, DiffItem{
				ResourceKind:   kind,
				Name:           name,
				LocalVersion:   localVersion,
				DesiredVersion: desiredVersion,
				DesiredMTime:   desiredMTimeByName[name],
				Kind:           KindUpdate,
			})
		}
	}
	return items
}
