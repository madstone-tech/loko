package cmd

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// TestCommandSurface pins the command set this release ships (FR-041, FR-043).
//
// It exists because the removals are easy to undo by accident: restoring a
// deleted command is a one-line AddCommand, and nothing else would notice. A
// failure here means either a command came back or a new one arrived without
// the list being updated deliberately.
func TestCommandSurface(t *testing.T) {
	t.Parallel()

	// "help" is absent: cobra adds it lazily at execute time, not at
	// registration, so it never appears in rootCmd.Commands() from a test.
	want := []string{"completion", "export", "fmt", "mcp", "validate", "version"}

	var got []string
	for _, c := range rootCmd.Commands() {
		got = append(got, c.Name())
	}
	sort.Strings(got)

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("command set changed.\n got: %v\nwant: %v", got, want)
	}
}

// TestRemovedCommandsStayRemoved names them individually so a failure says
// which one came back, and why it was taken out.
func TestRemovedCommandsStayRemoved(t *testing.T) {
	t.Parallel()

	reasons := map[string]string{
		"build":  "returns with the renderer stage; it drew from the deleted file-tree model",
		"serve":  "returns with the renderer stage",
		"watch":  "folds into serve in the renderer stage",
		"init":   "returns with the renderer stage; it scaffolded a loko.toml tree",
		"new":    "removed permanently — authoring is editing source, or the MCP write tools",
		"api":    "removed permanently (FR-041)",
		"import": "belongs to the observation-adapter stage",
		"diff":   "belongs to the diff stage",
	}

	present := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		present[c.Name()] = true
	}
	for name, why := range reasons {
		if present[name] {
			t.Errorf("%q is registered again; it was removed because it %s", name, why)
		}
	}
}

// TestNoDriftFlag covers FR-043: drift detection is not merely unimplemented,
// it is unreachable. The whole point of the release is that two sources of
// truth no longer exist to drift apart.
func TestNoDriftFlag(t *testing.T) {
	t.Parallel()

	for _, c := range rootCmd.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			if strings.Contains(f.Name, "drift") {
				t.Errorf("%s has a --%s flag; drift detection was removed", c.Name(), f.Name)
			}
		})
	}
}

// TestValidateFlags pins the validate surface described in contracts/cli.md.
func TestValidateFlags(t *testing.T) {
	t.Parallel()

	flags := validateCmd.Flags()
	for _, name := range []string{"strict", "format"} {
		if flags.Lookup(name) == nil {
			t.Errorf("validate is missing --%s", name)
		}
	}
	if flags.Lookup("exit-code") != nil {
		t.Error("validate still has --exit-code; exit codes are unconditional now")
	}
	if flags.Lookup("check-drift") != nil {
		t.Error("validate still has --check-drift")
	}
}

func TestFmtAndExportFlags(t *testing.T) {
	t.Parallel()

	if fmtCmd.Flags().Lookup("check") == nil {
		t.Error("fmt is missing --check")
	}
	for _, name := range []string{"format", "out"} {
		if exportCmd.Flags().Lookup(name) == nil {
			t.Errorf("export is missing --%s", name)
		}
	}
}
