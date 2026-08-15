package skills

import (
	"github.com/bruaguspons/devsync/core"
)

// LocalState walks ~/.claude/skills/* (p.installedRoot) and returns
// one core.Resource per installed skill directory, content-hashed via
// hashDirTree. Always a fresh walk — never reads the lock file.
func (p *Provider) LocalState() ([]core.Resource, error) {
	return loadSkillResources(p.installedRoot, "hash installed skill")
}
