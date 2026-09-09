# PowerShell completion for tcping
#
# Install by dot-sourcing it from your profile ($PROFILE):
#   . C:\path\to\tcping.ps1
#
# The sqlite3 output flag (--sqlite) is missing on purpose: it is not available
# in the Windows builds.

$script:TcpingFlags = [ordered]@{
	'-h'                     = 'Show the available flags and exit'
	'--version'              = 'Show the version and exit'
	'-u'                     = 'Check for updates and exit'

	'-c'                     = 'Stop after <n> probes, regardless of the result'
	'-i'                     = 'Interval between probes, in seconds'
	'-t'                     = 'Time to wait for a response, in seconds. 0 means infinite'
	'-4'                     = 'Only use IPv4 to initiate probes'
	'-6'                     = 'Only use IPv6 to initiate probes'
	'-I'                     = 'Interface name or IP address to initiate the probes from'

	'-r'                     = 'Retry resolving the hostname after <n> failed probes'
	'--resolve-every-probe'  = 'Resolve the hostname before every probe instead of only at startup'
	'--dns-server'           = 'Custom DNS server IP to use, e.g. 1.1.1.1:53'
	'--dns-timeout'          = 'Time to wait for a DNS response, in seconds. 0 means infinite'

	'-D'                     = 'Show a timestamp for each probe'
	'--no-color'             = 'Do not colorize the output'
	'--show-source-address'  = 'Show the source address and port used for the probes'
	'--failures-only'        = 'Only show the failed probes'
	'--no-stats'             = 'Do not show the statistics when the program exits'
	'-v'                     = 'Show all the details an HTTP(S) or UDP probe collects'

	'-j'                     = 'Output in JSON format'
	'--pretty'               = 'Prettify the JSON output. No effect without -j'
	'--csv'                  = 'Store the output in a CSV file'
	'--csv-fixed-name'       = 'Use the --csv filename as it is, without a date/time suffix'

	'--alloy'                = 'Send the results to a Grafana Alloy OTLP HTTP endpoint'
	'--influxdb'             = 'Write the results to an InfluxDB v2 or v3 server'
	'--influxdb-org'         = 'InfluxDB organization to write to'
	'--influxdb-bucket'      = 'InfluxDB bucket to write to'
	'--influxdb-token'       = 'InfluxDB API token'
	'--stats-interval'       = 'How often to send the statistics to Alloy or InfluxDB, in seconds'
	'--source-label'         = 'Name this machine in the metrics sent to Alloy or InfluxDB'

	'--insecure'             = 'Do not verify the server certificate of an https:// target'
	'--udp-server'           = 'Do not probe. Echo every received UDP datagram back to its sender'
}

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
