# PowerShell completion for tcping
#
# Install by loading it from your profile ($PROFILE):
#   tcping --completions powershell | Out-String | Invoke-Expression
#
# The sqlite3 output flag (--sqlite) is missing on purpose: it is not available
# in the Windows builds.

Register-ArgumentCompleter -Native -CommandName tcping, tcping.exe -ScriptBlock {
	param($wordToComplete, $commandAst, $cursorPosition)

	# Everything that is not a flag is the target or its port, which we
	# cannot guess.
	if (-not $wordToComplete.StartsWith('-')) {
		return
	}

	# PowerShell hash tables ignore case, which would make -i and -I the same
	# key, so the flags are kept in a dictionary that does not.
	$flags = [System.Collections.Specialized.OrderedDictionary]::new([System.StringComparer]::Ordinal)
	$flags['-h'] = 'Show the available flags and exit'
	$flags['--version'] = 'Show the version and exit'
	$flags['-u'] = 'Check for updates and exit'
	$flags['--completions'] = 'Print the completion script for a shell and exit'

	$flags['-c'] = 'Stop after <n> probes, regardless of the result'
	$flags['-i'] = 'Interval between probes, in seconds'
	$flags['-t'] = 'Time to wait for a response, in seconds. 0 means infinite'
	$flags['-4'] = 'Only use IPv4 to initiate probes'
	$flags['-6'] = 'Only use IPv6 to initiate probes'
	$flags['-I'] = 'Interface name or IP address to initiate the probes from'

	$flags['-r'] = 'Retry resolving the hostname after <n> failed probes'
	$flags['--resolve-every-probe'] = 'Resolve the hostname before every probe instead of only at startup'
	$flags['--dns-server'] = 'Custom DNS server IP to use, e.g. 1.1.1.1:53'
	$flags['--dns-timeout'] = 'Time to wait for a DNS response, in seconds. 0 means infinite'

	$flags['-D'] = 'Show a timestamp for each probe'
	$flags['--no-color'] = 'Do not colorize the output'
	$flags['--show-source-address'] = 'Show the source address and port used for the probes'
	$flags['--failures-only'] = 'Only show the failed probes'
	$flags['--no-stats'] = 'Do not show the statistics when the program exits'
	$flags['-v'] = 'Show all the details an HTTP(S) or UDP probe collects'

	$flags['-j'] = 'Output in JSON format'
	$flags['--pretty'] = 'Prettify the JSON output. No effect without -j'
	$flags['--json-url'] = 'Send the JSON output to an HTTP server instead of printing it'
	$flags['--json-server'] = 'Do not probe. Print the JSON events posted to the given host and port'
	$flags['--csv'] = 'Store the output in a CSV file'
	$flags['--csv-fixed-name'] = 'Use the --csv filename as it is, without a date/time suffix'

	$flags['--otlp'] = 'Send the results to an OTLP HTTP endpoint'
	$flags['--otlp-header'] = 'Extra HTTP header for the OTLP endpoint, as "Name: value"'
	$flags['--influxdb'] = 'Write the results to an InfluxDB v2 or v3 server'
	$flags['--influxdb-org'] = 'InfluxDB organization to write to'
	$flags['--influxdb-bucket'] = 'InfluxDB bucket to write to'
	$flags['--influxdb-token'] = 'InfluxDB API token'
	$flags['--stats-interval'] = 'How often to send the statistics to OTLP or InfluxDB, in seconds'
	$flags['--source-label'] = 'Name this machine in the metrics sent to OTLP or InfluxDB'

	$flags['--insecure'] = 'Do not verify the server certificate of an https:// target'
	$flags['--proxy'] = 'Send TCP and HTTP(S) probes through a proxy, e.g. http://proxy.example.com:3128'
	$flags['--udp-server'] = 'Do not probe. Echo every received UDP datagram back to its sender'

	$flags.GetEnumerator() |
		Where-Object { $_.Key -clike "$wordToComplete*" } |
		ForEach-Object {
			[System.Management.Automation.CompletionResult]::new(
				$_.Key, $_.Key, 'ParameterName', $_.Value)
		}
}
