package cli

import (
	"fmt"
	"os"

	"github.com/pouriyajamshidi/tcping/v3/internal/completions"
)

// completionScript returns the completion script for the named shell.
func completionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return completions.Bash, nil
	case "zsh":
		return completions.Zsh, nil
	case "fish":
		return completions.Fish, nil
	case "powershell":
		return completions.PowerShell, nil
	default:
		return "", fmt.Errorf("no completions for %q, use bash, zsh, fish or powershell", shell)
	}
}

// printCompletions prints the completion script for the named shell and
// exits.
func printCompletions(shell string) {
	script, err := completionScript(shell)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Print(script)
	os.Exit(0)
}
