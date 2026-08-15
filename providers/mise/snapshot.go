package mise

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bruaguspons/devsync/core"
)

// ErrSnapshotMissing is returned when the cached mise.toml snapshot does
// not exist locally. Callers must report this clearly and MUST NOT fall
// back to a live network fetch.
var ErrSnapshotMissing = errors.New("no cached mise.toml snapshot found; run the devsync installer to fetch one")

// DefaultSnapshotPath is the on-disk location of the cached mise.toml
// snapshot, mirroring mise's own XDG-data convention.
func DefaultSnapshotPath() string {
	return core.UserHomeDir() + "/.local/share/devsync/mise.toml.snapshot"
}

// rawToolsFile is the minimal shape of a mise.toml file needed for
// decoding the [tools] table. Values may be a plain string ("latest"),
// a table ({version = "latest"}), or (in mise's grammar) an array of
// strings/tables for multi-version installs; toml.Primitive defers
// exact typing to decodeToolVersion.
type rawToolsFile struct {
	Tools map[string]toml.Primitive `toml:"tools"`
}

// DesiredState reads and decodes the cached mise.toml snapshot into
// normalized core.Resource entries. It never performs a network fetch;
// if the snapshot file is absent it returns ErrSnapshotMissing.
func (p *Provider) DesiredState() ([]core.Resource, error) {
	path := p.snapshotPath
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSnapshotMissing
		}
		return nil, fmt.Errorf("stat snapshot %q: %w", path, err)
	}

	var raw rawToolsFile
	meta, err := toml.DecodeFile(path, &raw)
	if err != nil {
		return nil, fmt.Errorf("decode snapshot %q: %w", path, err)
	}

	states := make([]core.Resource, 0, len(raw.Tools))
	for name, prim := range raw.Tools {
		version, err := decodeToolVersion(meta, prim)
		if err != nil {
			return nil, fmt.Errorf("decode tool %q in snapshot: %w", name, err)
		}
		states = append(states, core.Resource{
			Kind:              core.ResourceKindTool,
			Name:              normalizeToolName(name),
			Version:           version,
			ComparisonVersion: version,
		})
	}
	return states, nil
}

// decodeToolVersion extracts a version string from a [tools] value that
// may be a plain string ("latest"), a table ({version = "latest"}), or an
// array (multi-version install — first entry wins for diff purposes).
func decodeToolVersion(meta toml.MetaData, prim toml.Primitive) (string, error) {
	// Plain string form: tool = "latest"
	var asString string
	if err := meta.PrimitiveDecode(prim, &asString); err == nil && asString != "" {
		return asString, nil
	}

	// Table form: tool = { version = "latest" }
	var asTable struct {
		Version string `toml:"version"`
	}
	if err := meta.PrimitiveDecode(prim, &asTable); err == nil && asTable.Version != "" {
		return asTable.Version, nil
	}

	// Array form: tool = ["latest"] or [{ version = "latest" }, ...]
	var asStringArray []string
	if err := meta.PrimitiveDecode(prim, &asStringArray); err == nil && len(asStringArray) > 0 {
		return asStringArray[0], nil
	}
	var asTableArray []struct {
		Version string `toml:"version"`
	}
	if err := meta.PrimitiveDecode(prim, &asTableArray); err == nil && len(asTableArray) > 0 && asTableArray[0].Version != "" {
		return asTableArray[0].Version, nil
	}

	return "", fmt.Errorf("unrecognized [tools] value shape")
}
