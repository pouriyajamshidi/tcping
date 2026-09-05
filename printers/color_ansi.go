package printers

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// The escape prefixes of the colors we print with. They are all from the
// basic 16, which every terminal we ship to renders the same way.
const (
	cyanCode        = "\033[36m"
	lightCyanCode   = "\033[96m"
	greenCode       = "\033[32m"
	lightGreenCode  = "\033[92m"
	yellowCode      = "\033[33m"
	lightYellowCode = "\033[93m"
	redCode         = "\033[31m"
	lightBlueCode   = "\033[94m"
	resetCode       = "\033[0m"
)

// colorEnabled says whether we emit escape codes at all. It is off when
// stdout is not a terminal, so redirecting tcping to a file gives plain text,
// and when NO_COLOR is set. FORCE_COLOR turns it back on for terminals we
// cannot detect, like mintty in Git Bash on Windows.
var colorEnabled = os.Getenv("FORCE_COLOR") != "" ||
	(os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd())))

// ansi is the escape prefix of one color.
type ansi string

// Printf writes the text in the color and resets right after, so whatever is
// printed next is not colored too. Callers hand it optional parts that can
// come out empty, and there is nothing to color in that case.
func (c ansi) Printf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if msg == "" {
		return
	}

	if !colorEnabled {
		fmt.Fprint(os.Stdout, msg)
		return
	}

	fmt.Fprint(os.Stdout, string(c)+msg+resetCode)
}

// Color function aliases to use when printing information
var (
	printCyan        = ansi(cyanCode).Printf
	printLightCyan   = ansi(lightCyanCode).Printf
	printGreen       = ansi(greenCode).Printf
	printLightGreen  = ansi(lightGreenCode).Printf
	printYellow      = ansi(yellowCode).Printf
	printLightYellow = ansi(lightYellowCode).Printf
	printRed         = ansi(redCode).Printf
	printLightBlue   = ansi(lightBlueCode).Printf
)
