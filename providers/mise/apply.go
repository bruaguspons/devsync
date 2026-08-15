package mise

import (
	"fmt"
	"os/exec"

	"github.com/bruaguspons/devsync/core"
)

// execCommand builds an *exec.Cmd for one DiffItem. Extracted so tests
// can point it at a fake `mise` binary via PATH without touching
// production behavior.
var execCommand = exec.Command

// commandForItem returns the argv slice (never a shell string) mise
// invocation for one confirmed DiffItem. Tool names/versions are passed
// as discrete argv elements, so they can never break out into extra
// shell tokens even if they contain characters like ";" or spaces.
func commandForItem(item core.DiffItem) (name string, args []string, err error) {
	switch item.Kind {
	case core.KindNew, core.KindUpdate:
		return "mise", []string{"use", item.Name + "@" + item.DesiredVersion}, nil
	case core.KindRemoved:
		return "mise", []string{"uninstall", item.Name}, nil
	default:
		return "", nil, fmt.Errorf("apply: unknown diff kind %q for %q", item.Kind, item.Name)
	}
}

// Apply applies one confirmed DiffItem via a `mise use`/`mise
// uninstall` shell-out.
func (p *Provider) Apply(item core.DiffItem) error {
	name, args, err := commandForItem(item)
	if err != nil {
		return err
	}

	cmd := execCommand(name, args...)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return fmt.Errorf("%s %v: %w (output: %s)", name, args, runErr, string(out))
	}
	return nil
}
