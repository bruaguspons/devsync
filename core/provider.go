package core

// Provider is the seam that decouples the generic diff/select/apply
// orchestration in this package from any concrete resource backend
// (mise-managed tools, repo skills, or future backends). Each backend
// implements Provider as an adapter; core never imports a concrete
// provider package.
type Provider interface {
	// Kind returns this provider's stable ResourceKind, used to tag
	// every Resource/DiffItem it produces and to route Apply calls
	// back to it.
	Kind() ResourceKind

	// LocalState returns the live, currently-installed state. Always a
	// fresh query (mise ls --json / on-disk hash walk) — never reads
	// the lock file. This is the authoritative "what's actually there"
	// signal.
	LocalState() ([]Resource, error)

	// DesiredState returns the repo-declared target state (mise.toml
	// snapshot / repo skills/ folder), read from the local checkout —
	// never a network fetch.
	DesiredState() ([]Resource, error)

	// Apply performs one confirmed DiffItem's install/update/remove
	// action and returns only an error.
	Apply(item DiffItem) error
}
