// Package completions holds the shell completion scripts. They are built into
// the binary so that "tcping --completions <shell>" can print them, and the
// Linux packages and the Nix flake install the same files.
package completions

import _ "embed"

// Bash is the bash completion script.
//
//go:embed tcping.bash
var Bash string

// Zsh is the zsh completion script.
//
//go:embed _tcping
var Zsh string

// Fish is the fish completion script.
//
//go:embed tcping.fish
var Fish string

// PowerShell is the PowerShell completion script.
//
//go:embed tcping.ps1
var PowerShell string
