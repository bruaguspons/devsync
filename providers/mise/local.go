package mise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bruaguspons/devsync/core"
)

// ErrNotInstalled is returned (wrapped) by LocalState when the `mise`
// executable cannot be found on PATH, so callers can distinguish "mise
// isn't installed" from any other failure and show a friendly message.
var ErrNotInstalled = errors.New("mise: executable not found in PATH; install it with: curl https://mise.run | sh")

// normalizeToolName strips surrounding quotes (as can appear when a TOML
// key like "aqua:boyter/scc" is read back with quoting) while preserving
// backend prefixes (e.g. "aqua:boyter/scc") verbatim.
func normalizeToolName(name string) string {
	return strings.Trim(name, `"'`)
}

// miseListEntry mirrors one element of the arrays returned by
// `mise ls --json`, verified against a live `mise ls --json` run.
// Only the fields this tool needs are decoded; unknown fields are ignored.
type miseListEntry struct {
	Version          string `json:"version"`
	RequestedVersion string `json:"requested_version"`
	Installed        bool   `json:"installed"`
	Active           bool   `json:"active"`
}

// miseListOutput is the top-level shape of `mise ls --json`:
// a map of tool name -> list of version entries (usually one, but mise
// can report multiple versions per tool).
type miseListOutput map[string][]miseListEntry

// runMiseListJSON invokes `mise ls --json` and returns its raw stdout.
// Extracted as a seam so tests can substitute fixture bytes.
func runMiseListJSON() ([]byte, error) {
	cmd := exec.Command("mise", "ls", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrNotInstalled
		}
		return nil, fmt.Errorf("mise ls --json: %w (stderr: %s)", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// decodeMiseList parses `mise ls --json` output into a normalized
// []core.Resource, one entry per tool using its active (or first) version.
func decodeMiseList(data []byte) ([]core.Resource, error) {
	var out miseListOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode mise ls --json output: %w", err)
	}

	states := make([]core.Resource, 0, len(out))
	for name, entries := range out {
		if len(entries) == 0 {
			continue
		}
		chosen := entries[0]
		for _, e := range entries {
			if e.Active {
				chosen = e
				break
			}
		}
		states = append(states, core.Resource{
			Kind:              core.ResourceKindTool,
			Name:              normalizeToolName(name),
			Version:           chosen.Version,
			ComparisonVersion: chosen.RequestedVersion,
		})
	}
	return states, nil
}

// LocalState returns the current installed tool state by invoking
// `mise ls --json` as ground truth.
func (p *Provider) LocalState() ([]core.Resource, error) {
	data, err := runMiseListJSON()
	if err != nil {
		return nil, err
	}
	return decodeMiseList(data)
}
