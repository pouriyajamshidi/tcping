package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/pouriyajamshidi/tcping/v3/internal/version"
)

// usage prints how tcping should be run
func usage() {
	fmt.Printf("\nTCPING version %s\n\n", version.Current)
	fmt.Println("Try running tcping like:")
	fmt.Println("tcping www.example.com 443")
	fmt.Println("Or use the <hostname/ip:port> format:")
	fmt.Println("tcping www.example.com:443")
	fmt.Println("Or probe over HTTP(S) by giving a URL:")
	fmt.Println("tcping https://www.example.com/health")
	fmt.Println("Or probe over UDP:")
	fmt.Println("tcping udp://www.example.com 53")
	fmt.Printf("\n[optional flags]\n")

	flag.VisitAll(func(f *flag.Flag) {
		flagName := f.Name
		if len(f.Name) > 1 {
			flagName = "-" + flagName
		}

		fmt.Printf("  -%s : %s\n", flagName, f.Usage)
	})

	os.Exit(1)
}

// showVersion displays the version and exits
func showVersion() {
	fmt.Printf("TCPING version %s\n", version.Current)
	os.Exit(0)
}
