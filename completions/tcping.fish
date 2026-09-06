# fish completion for tcping
#
# Install with:
#   install -Dm 644 tcping.fish ~/.config/fish/completions/tcping.fish

# The target is a hostname, an IP or a URL, and its port is a number.
# Neither is a file, so turn file completion off and offer known hosts.
complete -c tcping -f -a '(__fish_print_hostnames)' -d Target

complete -c tcping -s h -d 'Show the available flags and exit'
complete -c tcping -l version -d 'Show the version and exit'
complete -c tcping -s u -d 'Check for updates and exit'

complete -c tcping -s c -x -d 'Stop after <n> probes, regardless of the result'
complete -c tcping -s i -x -d 'Interval between probes, in seconds'
complete -c tcping -s t -x -d 'Time to wait for a response, in seconds. 0 means infinite'
complete -c tcping -s 4 -d 'Only use IPv4 to initiate probes'
complete -c tcping -s 6 -d 'Only use IPv6 to initiate probes'
complete -c tcping -s I -x -a '(__fish_print_interfaces)' -d 'Interface name or IP address to initiate the probes from'

complete -c tcping -s r -x -d 'Retry resolving the hostname after <n> failed probes'
complete -c tcping -l resolve-every-probe -d 'Resolve the hostname before every probe instead of only at startup'
complete -c tcping -l dns-server -x -d 'Custom DNS server IP to use, e.g. 1.1.1.1:53'
complete -c tcping -l dns-timeout -x -d 'Time to wait for a DNS response, in seconds. 0 means infinite'

complete -c tcping -s D -d 'Show a timestamp for each probe'
complete -c tcping -l no-color -d 'Do not colorize the output'
complete -c tcping -l show-source-address -d 'Show the source address and port used for the probes'
complete -c tcping -l failures-only -d 'Only show the failed probes'
complete -c tcping -l no-stats -d 'Do not show the statistics when the program exits'
complete -c tcping -s v -d 'Show all the details an HTTP(S) or UDP probe collects'

complete -c tcping -s j -d 'Output in JSON format'
complete -c tcping -l pretty -d 'Prettify the JSON output. No effect without -j'
complete -c tcping -l csv -r -F -d 'Store the output in a CSV file'
complete -c tcping -l csv-fixed-name -d 'Use the --csv filename as it is, without a date/time suffix'
complete -c tcping -l db -r -F -d 'Store the output in a sqlite3 database'

complete -c tcping -l alloy -x -d 'Send the results to a Grafana Alloy OTLP HTTP endpoint'
complete -c tcping -l influxdb -x -d 'Write the results to an InfluxDB v2 or v3 server'
complete -c tcping -l influxdb-org -x -d 'InfluxDB organization to write to'
complete -c tcping -l influxdb-bucket -x -d 'InfluxDB bucket to write to'
complete -c tcping -l influxdb-token -x -d 'InfluxDB API token'
complete -c tcping -l stats-interval -x -d 'How often to send the statistics to Alloy or InfluxDB, in seconds'
complete -c tcping -l source-label -x -d 'Name this machine in the metrics sent to Alloy or InfluxDB'

complete -c tcping -l insecure -d 'Do not verify the server certificate of an https:// target'
complete -c tcping -l udp-server -d 'Do not probe. Echo every received UDP datagram back to its sender'
