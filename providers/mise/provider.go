// Package mise implements core.Provider for mise-managed CLI tools,
// preserving the pre-refactor behavior of the standalone mise-sync
// tool (mise.go/snapshot.go/apply.go) behind the generic core.Provider
// seam.
package mise

import "github.com/bruaguspons/devsync/core"

// Provider implements core.Provider for mise-managed tools.
type Provider struct {
	snapshotPath string
}

// New constructs a mise Provider. snapshotPath is the on-disk location
// of the cached mise.toml snapshot used as desired state.
func New(snapshotPath string) *Provider {
	return &Provider{snapshotPath: snapshotPath}
}

// Kind returns core.ResourceKindTool.
func (p *Provider) Kind() core.ResourceKind {
	return core.ResourceKindTool
}
