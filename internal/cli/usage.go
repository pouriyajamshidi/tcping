package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pouriyajamshidi/tcping/v3/internal/version"
)

// printUsage writes how tcping should be run to w.
func printUsage(w io.Writer) {
	fmt.Fprintf(w, "\nTCPING version %s\n\n", version.Current)
	fmt.Fprintln(w, "Try running tcping like:")
	fmt.Fprintln(w, "tcping www.example.com 443")
	fmt.Fprintln(w, "Or use the <hostname/ip:port> format:")
	fmt.Fprintln(w, "tcping www.example.com:443")
	fmt.Fprintln(w, "Or probe over HTTP(S) by giving a URL:")
	fmt.Fprintln(w, "tcping https://www.example.com/health")
	fmt.Fprintln(w, "Or probe over UDP:")
	fmt.Fprintln(w, "tcping udp://www.example.com 53")
	fmt.Fprint(w, "\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		fmt.Fprintf(w, "  -%s : %s\n", flagName, f.Usage)
	})
}

// usage is what a wrong command line gets, so it goes to stderr like the
// other errors and ends the program with a failure.
func usage() {
	printUsage(os.Stderr)
	os.Exit(1)
}

// help is what -h gets. Asking for it is not a mistake, so it goes to stdout,
// where it can be piped into a pager or grep, and ends the program normally.
func help() {
	printUsage(os.Stdout)
	os.Exit(0)
}

// showVersion displays the version and exits
func showVersion() {
	fmt.Printf("TCPING version %s\n", version.Current)
	os.Exit(0)
}
