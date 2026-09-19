package cli

import (
	"flag"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/pouriyajamshidi/tcping/v3/internal/completions"
)

// The completion scripts under internal/completions are written by hand, one per
// shell, so nothing stops them from falling behind when a flag is added,
// renamed or removed. The test below reads the flag names back out of each
// script and compares them to the flags the program really defines.

// -h is not one of our flags. The flag package handles it on its own, so the
// scripts offer it even though nothing registers it.
const helpFlag = "h"

// registeredFlagNames is the name of every flag tcping defines.
func registeredFlagNames(t *testing.T) map[string]bool {
	t.Helper()

	oldCommandLine := flag.CommandLine
	defer func() {
		flag.CommandLine = oldCommandLine
	}()

	flag.CommandLine = flag.NewFlagSet("completions", flag.ContinueOnError)
	registerFlags()

	names := map[string]bool{}
	flag.VisitAll(func(f *flag.Flag) {
		names[f.Name] = true
	})

	return names
}

// The flags the bash script offers are the words of its "flags" variable.
var bashFlagBlock = regexp.MustCompile(`(?s)flags="(.*?)"`)

func bashFlags(script string) []string {
	block := bashFlagBlock.FindStringSubmatch(script)
	if block == nil {
		return nil
	}

	names := []string{}
	for word := range strings.FieldsSeq(block[1]) {
		names = append(names, strings.TrimLeft(word, "-"))
	}

	return names
}

// Every line of the zsh _arguments call starts with the flag it describes,
// which may carry a list of the flags it excludes, as in '(-6)-4[...]'.
var zshFlagSpec = regexp.MustCompile(`(?m)^\s*'(?:\([^)]*\))?-{1,2}([\w-]+)=?\[`)

func zshFlags(script string) []string {
	names := []string{}
	for _, match := range zshFlagSpec.FindAllStringSubmatch(script, -1) {
		names = append(names, match[1])
	}

	return names
}

// fish names a short flag with -s and a long one with -l.
var fishFlagSpec = regexp.MustCompile(`\s-[sl]\s+(\S+)`)

func fishFlags(script string) []string {
	names := []string{}

	for line := range strings.SplitSeq(script, "\n") {
		// The descriptions are quoted and can hold anything, so only the
		// part of the line before them is searched.
		line, _, _ = strings.Cut(line, "'")

		for _, match := range fishFlagSpec.FindAllStringSubmatch(line, -1) {
			names = append(names, match[1])
		}
	}

	return names
}

// The PowerShell script adds its flags one per line, as in
// $script:TcpingFlags['-c'] = '...'.
var powershellFlagSpec = regexp.MustCompile(`(?m)^\$script:TcpingFlags\['-{1,2}([\w-]+)'\]\s*=`)

func powershellFlags(script string) []string {
	names := []string{}
	for _, match := range powershellFlagSpec.FindAllStringSubmatch(script, -1) {
		names = append(names, match[1])
	}

	return names
}

func TestCompletionScriptsMatchFlags(t *testing.T) {
	registered := registeredFlagNames(t)

	tests := []struct {
		name    string
		script  string
		flagsIn func(string) []string
		// Flags the script leaves out on purpose.
		omitted map[string]bool
	}{
		{
			name:    "bash",
			script:  completions.Bash,
			flagsIn: bashFlags,
		},
		{
			name:    "zsh",
			script:  completions.Zsh,
			flagsIn: zshFlags,
		},
		{
			name:    "fish",
			script:  completions.Fish,
			flagsIn: fishFlags,
		},
		{
			name:    "powershell",
			script:  completions.PowerShell,
			flagsIn: powershellFlags,
			// The sqlite3 printer is a stub on Windows, so the Windows
			// script does not offer --sqlite.
			omitted: map[string]bool{"sqlite": true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			completed := map[string]bool{}
			for _, name := range test.flagsIn(test.script) {
				completed[name] = true
			}

			if len(completed) == 0 {
				t.Fatalf("found no flags in the %s script, this test can no longer read it", test.name)
			}

			for _, name := range slices.Sorted(maps.Keys(registered)) {
				if !completed[name] && !test.omitted[name] {
					t.Errorf("-%s is missing from the %s script", name, test.name)
				}
			}

			for _, name := range slices.Sorted(maps.Keys(completed)) {
				if !registered[name] && name != helpFlag {
					t.Errorf("-%s is completed by the %s script but is not a tcping flag", name, test.name)
				}
			}

			for _, name := range slices.Sorted(maps.Keys(test.omitted)) {
				if completed[name] {
					t.Errorf("-%s is left out of the %s script on purpose, but it is completed there", name, test.name)
				}
			}
		})
	}
}

// --completions prints the scripts the release archives and packages ship,
// so what a user saves from the binary is the same file they would install
// by hand.
func TestCompletionScriptIsTheShippedFile(t *testing.T) {
	tests := []struct {
		shell string
		path  string
	}{
		{shell: "bash", path: "../completions/tcping.bash"},
		{shell: "zsh", path: "../completions/_tcping"},
		{shell: "fish", path: "../completions/tcping.fish"},
		{shell: "powershell", path: "../completions/tcping.ps1"},
	}

	for _, test := range tests {
		t.Run(test.shell, func(t *testing.T) {
			want, err := os.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}

			got, err := completionScript(test.shell)
			if err != nil {
				t.Fatal(err)
			}

			if got != string(want) {
				t.Errorf("--completions %s does not print %s", test.shell, test.path)
			}
		})
	}
}

func TestCompletionScriptRejectsAnUnknownShell(t *testing.T) {
	_, err := completionScript("tcsh")
	if err == nil {
		t.Fatal("completionScript(\"tcsh\") returned no error")
	}

	want := `no completions for "tcsh", use bash, zsh, fish or powershell`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err, want)
	}
}
