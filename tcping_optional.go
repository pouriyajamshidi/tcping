package main

import (
	"flag"
	"os"
)

type tcpingUserInput struct {
	useIPv4                   *bool
	useIPv6                   *bool
	retryHostnameResolveAfter *uint
	probesBeforeQuit          *uint
	outputJSON                *bool
	prettyJSON                *bool
	noColor                   *bool
	showTimestamp             *bool
	saveToCSV                 *string
	showVer                   *bool
	checkUpdates              *bool
	secondsBetweenProbes      *float64
	timeout                   *float64
	outputDB                  *string
	interfaceName             *string
	showSourceAddress         *bool
	showFailuresOnly          *bool
	showHelp                  *bool
	args                      []string
}

type OptionalTcping func(tcping *tcping, userInput *tcpingUserInput) error

func newProcessUserInput(opts ...OptionalTcping) *tcping {

	tcping := defaultProcessUserInput()
	userInput := defaultUserSettings()

	for _, opt := range opts {
		opt(tcping, &userInput)
	}

	return tcping
}

// Tcping Default Settings
func defaultProcessUserInput() *tcping {
	return &tcping{}
}

func defaultUserSettings() tcpingUserInput {

	// overlap tcping.go 340 ~ 357
	useIPv4 := flag.Bool("4", false, "only use IPv4.")
	useIPv6 := flag.Bool("6", false, "only use IPv6.")
	retryHostnameResolveAfter := flag.Uint("r", 0, "retry resolving target's hostname after <n> number of failed probes. e.g. -r 10 to retry after 10 failed probes.")
	probesBeforeQuit := flag.Uint("c", 0, "stop after <n> probes, regardless of the result. By default, no limit will be applied.")
	outputJSON := flag.Bool("j", false, "output in JSON format.")
	prettyJSON := flag.Bool("pretty", false, "use indentation when using json output format. No effect without the '-j' flag.")
	noColor := flag.Bool("no-color", false, "do not colorize output.")
	showTimestamp := flag.Bool("D", false, "show timestamp in output.")
	saveToCSV := flag.String("csv", "", "path and file name to store tcping output to CSV file...If user prompts for stats, it will be saved to a file with the same name and _stats appended.")
	showVer := flag.Bool("v", false, "show version.")
	checkUpdates := flag.Bool("u", false, "check for updates and exit.")
	secondsBetweenProbes := flag.Float64("i", 1, "interval between sending probes. Real number allowed with dot as a decimal separator. The default is one second")
	timeout := flag.Float64("t", 1, "time to wait for a response, in seconds. Real number allowed. 0 means infinite timeout.")
	outputDB := flag.String("db", "", "path and file name to store tcping output to sqlite database.")
	interfaceName := flag.String("I", "", "interface name or address.")
	showSourceAddress := flag.Bool("show-source-address", false, "Show source address and port used for probes.")
	showFailuresOnly := flag.Bool("show-failures-only", false, "Show only the failed probes.")
	showHelp := flag.Bool("h", false, "show help message.")

	flag.CommandLine.Usage = usage
	permuteArgs(os.Args[1:])
	flag.Parse()
	args := flag.Args()

	return tcpingUserInput{
		useIPv4:                   useIPv4,
		useIPv6:                   useIPv6,
		retryHostnameResolveAfter: retryHostnameResolveAfter,
		probesBeforeQuit:          probesBeforeQuit,
		outputJSON:                outputJSON,
		prettyJSON:                prettyJSON,
		noColor:                   noColor,
		showTimestamp:             showTimestamp,
		saveToCSV:                 saveToCSV,
		showVer:                   showVer,
		checkUpdates:              checkUpdates,
		secondsBetweenProbes:      secondsBetweenProbes,
		timeout:                   timeout,
		outputDB:                  outputDB,
		interfaceName:             interfaceName,
		showSourceAddress:         showSourceAddress,
		showFailuresOnly:          showFailuresOnly,
		showHelp:                  showHelp,
		args:                      args,
	}
}

// set printer
func withMustPrinter() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {

		setPrinter(tcping, userInput.outputJSON, userInput.prettyJSON, userInput.noColor, userInput.showTimestamp, userInput.showSourceAddress, userInput.outputDB, userInput.saveToCSV, userInput.args)
		return nil
	}
}

// set show version (-v)
func withMustShowVersion() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {

		if *userInput.showVer {
			showVersion(tcping)
		}
		return nil
	}
}

// set show help (-h)
func withMustShowHelp() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {
		if *userInput.showHelp {
			usage()
		}
		return nil
	}
}

// set check updates (-u)
func withMustCheckUpdates() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {
		if *userInput.checkUpdates {
			checkForUpdates(tcping)
		}
		return nil
	}
}

// host and port must be specified
func withMustSpecificHostAndPort() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {
		if len(userInput.args) < 2 {
			usage()
		}
		return nil
	}
}

// set ip flags
func withMustIPFlags() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {
		setIPFlags(tcping, userInput.useIPv4, userInput.useIPv6)
		return nil
	}
}

// set port
func withMustPort() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {
		setPort(tcping, userInput.args)
		return nil
	}
}

// set generic args
func withMustGenericArgs() OptionalTcping {
	return func(tcping *tcping, userInput *tcpingUserInput) error {

		genericArgs := genericUserInputArgs{
			retryResolve:         userInput.retryHostnameResolveAfter,
			probesBeforeQuit:     userInput.probesBeforeQuit,
			timeout:              userInput.timeout,
			secondsBetweenProbes: userInput.secondsBetweenProbes,
			intName:              userInput.interfaceName,
			showFailuresOnly:     userInput.showFailuresOnly,
			showSourceAddress:    userInput.showSourceAddress,
			args:                 userInput.args,
		}
		setGenericArgs(tcping, genericArgs)
		return nil
	}
}
