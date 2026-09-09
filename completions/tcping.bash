# bash completion for tcping
#
# Install with:
#   sudo install -Dm 644 tcping.bash /usr/share/bash-completion/completions/tcping

_tcping() {
	local cur prev opt flags

	cur=${COMP_WORDS[COMP_CWORD]}
	prev=${COMP_WORDS[COMP_CWORD - 1]}

	# tcping takes flags with either one or two dashes, so strip them
	# before looking the previous word up.
	opt=""
	if [[ $prev == -* ]]; then
		opt=${prev#-}
		opt=${opt#-}
	fi

	case $opt in
	csv | sqlite)
		mapfile -t COMPREPLY < <(compgen -f -- "$cur")
		compopt -o filenames
		return
		;;
	I)
		local interfaces
		interfaces=$(ls /sys/class/net 2>/dev/null || ifconfig -l 2>/dev/null)
		mapfile -t COMPREPLY < <(compgen -W "$interfaces" -- "$cur")
		return
		;;
	c | i | t | r | dns-server | dns-timeout | stats-interval | source-label | \
		alloy | influxdb | influxdb-org | influxdb-bucket | influxdb-token)
		# These need a value we cannot guess. Complete nothing rather
		# than offering flags where a value belongs.
		return
		;;
	esac

	flags="-h --version -u
		-c -i -t -4 -6 -I
		-r --resolve-every-probe --dns-server --dns-timeout
		-D --no-color --show-source-address --failures-only --no-stats -v
		-j --pretty --csv --csv-fixed-name --sqlite
		--alloy --influxdb --influxdb-org --influxdb-bucket --influxdb-token
		--stats-interval --source-label
		--insecure --udp-server"

	if [[ $cur == -* ]]; then
		mapfile -t COMPREPLY < <(compgen -W "$flags" -- "$cur")
		return
	fi

	# Anything else is the target or its port, which we cannot guess.
	mapfile -t COMPREPLY < <(compgen -A hostname -- "$cur")
}

complete -F _tcping tcping
