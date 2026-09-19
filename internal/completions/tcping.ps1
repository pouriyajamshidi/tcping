# PowerShell completion for tcping
#
# Install by loading it from your profile ($PROFILE):
#   tcping --completions powershell | Out-String | Invoke-Expression
#
# The sqlite3 output flag (--sqlite) is missing on purpose: it is not available
# in the Windows builds.

# PowerShell hash tables ignore case, which would make -i and -I the same
# key, so the flags are kept in a dictionary that does not.
$script:TcpingFlags = [System.Collections.Specialized.OrderedDictionary]::new([System.StringComparer]::Ordinal)
$script:TcpingFlags['-h'] = 'Show the available flags and exit'
$script:TcpingFlags['--version'] = 'Show the version and exit'
$script:TcpingFlags['-u'] = 'Check for updates and exit'
$script:TcpingFlags['--completions'] = 'Print the completion script for a shell and exit'

$script:TcpingFlags['-c'] = 'Stop after <n> probes, regardless of the result'
$script:TcpingFlags['-i'] = 'Interval between probes, in seconds'
$script:TcpingFlags['-t'] = 'Time to wait for a response, in seconds. 0 means infinite'
$script:TcpingFlags['-4'] = 'Only use IPv4 to initiate probes'
$script:TcpingFlags['-6'] = 'Only use IPv6 to initiate probes'
$script:TcpingFlags['-I'] = 'Interface name or IP address to initiate the probes from'

$script:TcpingFlags['-r'] = 'Retry resolving the hostname after <n> failed probes'
$script:TcpingFlags['--resolve-every-probe'] = 'Resolve the hostname before every probe instead of only at startup'
$script:TcpingFlags['--dns-server'] = 'Custom DNS server IP to use, e.g. 1.1.1.1:53'
$script:TcpingFlags['--dns-timeout'] = 'Time to wait for a DNS response, in seconds. 0 means infinite'

$script:TcpingFlags['-D'] = 'Show a timestamp for each probe'
$script:TcpingFlags['--no-color'] = 'Do not colorize the output'
$script:TcpingFlags['--show-source-address'] = 'Show the source address and port used for the probes'
$script:TcpingFlags['--failures-only'] = 'Only show the failed probes'
$script:TcpingFlags['--no-stats'] = 'Do not show the statistics when the program exits'
$script:TcpingFlags['-v'] = 'Show all the details an HTTP(S) or UDP probe collects'

$script:TcpingFlags['-j'] = 'Output in JSON format'
$script:TcpingFlags['--pretty'] = 'Prettify the JSON output. No effect without -j'
$script:TcpingFlags['--json-url'] = 'Send the JSON output to an HTTP server instead of printing it'
$script:TcpingFlags['--json-server'] = 'Do not probe. Print the JSON events posted to the given host and port'
$script:TcpingFlags['--csv'] = 'Store the output in a CSV file'
$script:TcpingFlags['--csv-fixed-name'] = 'Use the --csv filename as it is, without a date/time suffix'

$script:TcpingFlags['--otlp'] = 'Send the results to an OTLP HTTP endpoint'
$script:TcpingFlags['--otlp-header'] = 'Extra HTTP header for the OTLP endpoint, as "Name: value"'
$script:TcpingFlags['--influxdb'] = 'Write the results to an InfluxDB v2 or v3 server'
$script:TcpingFlags['--influxdb-org'] = 'InfluxDB organization to write to'
$script:TcpingFlags['--influxdb-bucket'] = 'InfluxDB bucket to write to'
$script:TcpingFlags['--influxdb-token'] = 'InfluxDB API token'
$script:TcpingFlags['--stats-interval'] = 'How often to send the statistics to OTLP or InfluxDB, in seconds'
$script:TcpingFlags['--source-label'] = 'Name this machine in the metrics sent to OTLP or InfluxDB'

$script:TcpingFlags['--insecure'] = 'Do not verify the server certificate of an https:// target'
$script:TcpingFlags['--udp-server'] = 'Do not probe. Echo every received UDP datagram back to its sender'

Register-ArgumentCompleter -Native -CommandName tcping, tcping.exe -ScriptBlock {
	param($wordToComplete, $commandAst, $cursorPosition)

	# Everything that is not a flag is the target or its port, which we
	# cannot guess.
	if (-not $wordToComplete.StartsWith('-')) {
		return
	}

	$script:TcpingFlags.GetEnumerator() |
		Where-Object { $_.Key -like "$wordToComplete*" } |
		ForEach-Object {
			[System.Management.Automation.CompletionResult]::new(
				$_.Key, $_.Key, 'ParameterName', $_.Value)
		}
}
