package core

import "time"

// ResourceKind identifies which provider owns a Resource/DiffItem —
// "tool" (mise-managed) or "skill" (repo skills/ folder).
type ResourceKind string

const (
	ResourceKindTool  ResourceKind = "tool"
	ResourceKindSkill ResourceKind = "skill"
)

// Resource represents a single tool or skill's installed or desired
// version, keyed by its normalized name (e.g. "aqua:boyter/scc" for
// tools, or a skill directory name for skills). For skills, Version
// holds the content hash (hex string), not a semver.
type Resource struct {
	Kind    ResourceKind
	Name    string
	Version string

	// ComparisonVersion, when non-empty on BOTH sides of a Diff() call,
	// is used instead of Version to decide equality (e.g. mise's raw
	// mise.toml declaration vs. requested_version) — Version itself
	// always stays the human-facing (resolved) value shown in the UI.
	ComparisonVersion string

	// SourceMTime is the max file mtime backing this resource,
	// populated only by the skills provider (cheap os.Stat walk, no
	// content read). Used exclusively by reconcileWithLock to decide
	// whether a cached hash can be reused; nil for tools and unused
	// otherwise.
	SourceMTime *time.Time
}

// DiffKind classifies a DiffItem relative to local vs. desired state.
// It is orthogonal to ResourceKind: DiffKind classifies the *action*,
// ResourceKind classifies the *domain*. The two are never confused
// because they are distinct Go types.
type DiffKind string

const (
	KindNew     DiffKind = "new"     // present in desired state, not installed locally
	KindRemoved DiffKind = "removed" // installed locally, not present in desired state
	KindUpdate  DiffKind = "update"  // present on both sides with differing versions
)

// DiffItem is one actionable row in the diff between local and desired
// state for a single resource.
type DiffItem struct {
	ResourceKind   ResourceKind
	Name           string
	LocalVersion   string // "" if not locally installed
	DesiredVersion string // "" if not in desired state
	Kind           DiffKind

	// DesiredMTime carries the desired-state Resource's SourceMTime
	// through to ApplyResult, so a successful apply can record it in
	// the lock file for future reconcileWithLock short-circuiting.
	// Skills only; nil for tools.
	DesiredMTime *time.Time

	// OwnershipWarning is true when this item is a foreign resource
	// (not lock-tracked) that shares a name with desired state but
	// differs in content — apply will overwrite unmanaged local
	// content. Set only by core.Run(), never by Diff().
	OwnershipWarning bool
}

// ApplyResult is the outcome of attempting to apply one DiffItem.
type ApplyResult struct {
	Item DiffItem
	Err  error
}
