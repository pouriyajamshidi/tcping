package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/pouriyajamshidi/tcping/v3/internal/version"
)

// usage prints how tcping should be run. It is only shown when the command
// line is wrong, so it goes to stderr like the other errors.
func usage() {
	fmt.Fprintf(os.Stderr, "\nTCPING version %s\n\n", version.Current)
	fmt.Fprintln(os.Stderr, "Try running tcping like:")
	fmt.Fprintln(os.Stderr, "tcping www.example.com 443")
	fmt.Fprintln(os.Stderr, "Or use the <hostname/ip:port> format:")
	fmt.Fprintln(os.Stderr, "tcping www.example.com:443")
	fmt.Fprintln(os.Stderr, "Or probe over HTTP(S) by giving a URL:")
	fmt.Fprintln(os.Stderr, "tcping https://www.example.com/health")
	fmt.Fprintln(os.Stderr, "Or probe over UDP:")
	fmt.Fprintln(os.Stderr, "tcping udp://www.example.com 53")
	fmt.Fprint(os.Stderr, "\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		fmt.Fprintf(os.Stderr, "  -%s : %s\n", flagName, f.Usage)
	})

	os.Exit(1)
}

// showVersion displays the version and exits
func showVersion() {
	fmt.Printf("TCPING version %s\n", version.Current)
	os.Exit(0)
}
