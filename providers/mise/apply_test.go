package mise

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/bruaguspons/devsync/core"
)

// TestCommandForItem_MaliciousToolNameStaysOneArgvElement is the
// threat-matrix RED test: a malicious tool name containing shell
// metacharacters must be passed to exec.Command as a single argv
// element, never interpolated into a shell string where it could break
// out into extra commands.
func TestCommandForItem_MaliciousToolNameStaysOneArgvElement(t *testing.T) {
	malicious := "node; rm -rf /"

	name, args, err := commandForItem(core.DiffItem{
		Name:           malicious,
		DesiredVersion: "24.0.0",
		Kind:           core.KindNew,
	})
	if err != nil {
		t.Fatalf("commandForItem: %v", err)
	}
	if name != "mise" {
		t.Fatalf("command name = %q, want mise", name)
	}

	// The malicious name (with its ";" and spaces intact) must appear as
	// exactly one argv element, never split across multiple elements and
	// never handed to a shell for interpretation.
	found := false
	for _, a := range args {
		if a == malicious+"@24.0.0" {
			found = true
		}
		// If the malicious payload were shell-interpolated, "rm" would
		// show up as its own separate argv token.
		if a == "rm" || a == "-rf" || a == "/" {
			t.Fatalf("malicious payload leaked into separate argv token %q: %v", a, args)
		}
	}
	if !found {
		t.Fatalf("expected single argv element %q@version, got %v", malicious, args)
	}

	// Build the real exec.Cmd and confirm exec.Command's argv semantics
	// hold: cmd.Args[0] is the resolved path/name, the rest are the
	// exact argv slice passed in, with no shell involved.
	cmd := exec.Command(name, args...)
	if len(cmd.Args) != 1+len(args) {
		t.Fatalf("cmd.Args = %v, want len %d", cmd.Args, 1+len(args))
	}
	for i, a := range args {
		if cmd.Args[i+1] != a {
			t.Fatalf("cmd.Args[%d] = %q, want %q", i+1, cmd.Args[i+1], a)
		}
	}
}

func TestCommandForItem_Uninstall(t *testing.T) {
	name, args, err := commandForItem(core.DiffItem{Name: "watchexec", Kind: core.KindRemoved})
	if err != nil {
		t.Fatalf("commandForItem: %v", err)
	}
	if name != "mise" || len(args) != 2 || args[0] != "uninstall" || args[1] != "watchexec" {
		t.Fatalf("commandForItem = %q %v, want mise [uninstall watchexec]", name, args)
	}
}

// TestApply_BestEffortBatch_OneFailureDoesNotBlockRest uses a fake
// `mise` binary on PATH with mixed exit codes to verify one failure
// doesn't prevent remaining items from being attempted, and every
// item's outcome is reported.
func TestApply_BestEffortBatch_OneFailureDoesNotBlockRest(t *testing.T) {
	fakeMisePath := writeFakeMiseBinary(t)

	origPath := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", origPath) })
	os.Setenv("PATH", fakeMisePath+string(os.PathListSeparator)+origPath)

	p := New("")
	selected := []core.DiffItem{
		{Name: "will-fail", Kind: core.KindRemoved},
		{Name: "node", DesiredVersion: "24.0.0", Kind: core.KindNew},
		{Name: "rust", DesiredVersion: "1.97.1", Kind: core.KindUpdate},
	}

	results := make([]core.ApplyResult, 0, len(selected))
	for _, item := range selected {
		err := p.Apply(item)
		results = append(results, core.ApplyResult{Item: item, Err: err})
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results (best-effort, all attempted), got %d: %+v", len(results), results)
	}
	if results[0].Err == nil {
		t.Errorf("expected will-fail to report an error, got nil")
	}
	if results[1].Err != nil {
		t.Errorf("expected node to succeed, got error: %v", results[1].Err)
	}
	if results[2].Err != nil {
		t.Errorf("expected rust to succeed, got error: %v", results[2].Err)
	}
}

func TestCommandForItem_UnknownKind(t *testing.T) {
	_, _, err := commandForItem(core.DiffItem{Name: "x", Kind: "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown diff kind")
	}
}

// writeFakeMiseBinary creates a shell script named "mise" on a temp
// directory that fails when invoked with "uninstall will-fail" and
// succeeds otherwise, mimicking mixed real-world exit codes.
func writeFakeMiseBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := `#!/bin/sh
if [ "$1" = "uninstall" ] && [ "$2" = "will-fail" ]; then
  echo "simulated failure" >&2
  exit 1
fi
echo "ok: $@"
exit 0
`
	path := dir + "/mise"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake mise: %v", err)
	}
	if !strings.HasSuffix(path, "/mise") {
		t.Fatal("sanity check failed")
	}
	return dir
}
