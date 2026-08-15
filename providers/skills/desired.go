package skills

import (
	"github.com/bruaguspons/devsync/core"
)

// DesiredState walks <repoRoot>/skills/* (p.desiredRoot) and returns one
// core.Resource per declared skill directory, content-hashed via
// hashDirTree. Reads from the local checkout only — never a network
// fetch.
func (p *Provider) DesiredState() ([]core.Resource, error) {
	return loadSkillResources(p.desiredRoot, "hash repo skill")
}
