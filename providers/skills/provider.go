package skills

import (
	"path/filepath"

	"github.com/bruaguspons/devsync/core"
)

// Provider implements core.Provider for repo-declared Claude Code
// skills.
type Provider struct {
	desiredRoot   string // desired state source, e.g. <checkout>/skills or the installed skills.snapshot dir
	installedRoot string // local state destination, typically ~/.claude/skills
}

// New constructs a skills Provider. desiredRoot is the directory
// containing one subdirectory per declared skill (the repo's skills/
// folder in a checkout, or the installed skills.snapshot/ bundle for
// a released binary); installedRoot is the destination directory
// (typically ~/.claude/skills).
func New(desiredRoot, installedRoot string) *Provider {
	return &Provider{
		desiredRoot:   desiredRoot,
		installedRoot: installedRoot,
	}
}

// Kind returns core.ResourceKindSkill.
func (p *Provider) Kind() core.ResourceKind {
	return core.ResourceKindSkill
}

// DefaultInstalledRoot returns ~/.claude/skills, the on-disk
// destination for installed skills.
func DefaultInstalledRoot() string {
	return filepath.Join(core.UserHomeDir(), ".claude", "skills")
}

// DefaultDesiredRoot returns the on-disk location of the bundled
// skills.snapshot directory, mirroring the mise provider's
// DefaultSnapshotPath() XDG-style convention.
func DefaultDesiredRoot() string {
	return filepath.Join(core.UserHomeDir(), ".local", "share", "devsync", "skills.snapshot")
}
