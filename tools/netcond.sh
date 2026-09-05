#!/usr/bin/env bash
#
# netcond.sh - apply temporary network conditions toward one destination so
# tcping can be tested against them.
#
# Everything it adds is tagged, so "clear" only removes what this script made
# and leaves the rest of your firewall alone.
#
# Needs root, iptables and tc (iproute2). IPv4 only.

set -euo pipefail

TAG="netcond"

usage() {
	cat <<'EOF'
Usage: netcond.sh <command> [args]

Commands:
  block <dst> [seconds]                 drop all traffic to dst
  delay <dst> <time> [jitter] [seconds] add latency to dst, e.g. 800ms 200ms
  loss  <dst> <percent> [seconds]       drop a share of packets to dst
  status                                show what is currently applied
  clear                                 remove everything this script added

If seconds is left out the condition stays until you run "clear".
If seconds is given the script waits, then cleans up, and Ctrl-C cleans up too.

Examples:
  sudo ./netcond.sh block 1.1.1.1 10
  sudo ./netcond.sh delay example.com 800ms 200ms 30
  sudo ./netcond.sh loss 9.9.9.9 25
  sudo ./netcond.sh clear
EOF
}

die() {
	echo "netcond: $*" >&2
	exit 1
}

require_root() {
	[[ $EUID -eq 0 ]] || die "needs root, run it with sudo"
}

# Turn a hostname or IP into a single IPv4 address.
resolve() {
	local host=$1 ip
	ip=$(getent ahostsv4 "$host" | awk 'NR==1 {print $1}')
	[[ -n $ip ]] || die "cannot resolve $host to an IPv4 address"
	echo "$ip"
}

# The interface the kernel would use to reach an address.
iface_for() {
	local ip=$1 dev
	dev=$(ip -4 route get "$ip" | awk '{for (i=1; i<NF; i++) if ($i == "dev") print $(i+1)}')
	[[ -n $dev ]] || die "no route to $ip"
	echo "$dev"
}

# Sleep for the given seconds, then let the EXIT trap clean up.
wait_then_clear() {
	local seconds=$1
	trap clear_all EXIT INT TERM
	echo "holding for ${seconds}s, Ctrl-C to stop early"
	sleep "$seconds"
}

cmd_block() {
	local dst=${1:-} seconds=${2:-}
	[[ -n $dst ]] || die "block needs a destination"

	local ip
	ip=$(resolve "$dst")

	iptables -I OUTPUT 1 -d "$ip" -m comment --comment "$TAG" -j DROP
	echo "blocking $ip"

	if [[ -n $seconds ]]; then
		wait_then_clear "$seconds"
	fi
}

cmd_delay() {
	local dst=${1:-} time=${2:-} jitter=${3:-} seconds=${4:-}
	[[ -n $dst && -n $time ]] || die "delay needs a destination and a time, e.g. 200ms"

	# jitter is optional, so a bare number in its place is really the duration
	if [[ $jitter =~ ^[0-9]+$ ]]; then
		seconds=$jitter
		jitter=""
	fi

	if [[ -n $jitter ]]; then
		apply_netem "$dst" delay "$time" "$jitter" distribution normal
	else
		apply_netem "$dst" delay "$time"
	fi

	if [[ -n $seconds ]]; then
		wait_then_clear "$seconds"
	fi
}

cmd_loss() {
	local dst=${1:-} percent=${2:-} seconds=${3:-}
	[[ -n $dst && -n $percent ]] || die "loss needs a destination and a percentage"

	apply_netem "$dst" loss "${percent%\%}%"

	if [[ -n $seconds ]]; then
		wait_then_clear "$seconds"
	fi
}

# Send traffic for one destination into its own band and hang netem off it, so
# only that destination is affected and everything else keeps its normal path.
apply_netem() {
	local dst=$1
	shift

	local ip iface
	ip=$(resolve "$dst")
	iface=$(iface_for "$ip")

	tc qdisc show dev "$iface" | grep -q "qdisc prio 1:" ||
		tc qdisc add dev "$iface" root handle 1: prio

	tc qdisc replace dev "$iface" parent 1:3 handle 30: netem "$@"

	tc filter replace dev "$iface" protocol ip parent 1:0 prio 1 u32 \
		match ip dst "$ip"/32 \
		flowid 1:3

	echo "applied netem $* to $ip on $iface"
}

cmd_status() {
	echo "iptables rules:"
	iptables -S OUTPUT | grep -- "--comment \"$TAG\"" || echo "  none"

	echo
	echo "netem qdiscs:"
	tc qdisc show | grep netem || echo "  none"
}

clear_all() {
	trap - EXIT INT TERM

	# Drop our tagged rules one at a time, newest first.
	local line
	while line=$(iptables -L OUTPUT --line-numbers -n | awk -v tag="$TAG" '$0 ~ tag {print $1; exit}'); [[ -n $line ]]; do
		iptables -D OUTPUT "$line"
	done

	# Remove the root qdisc on every interface that ended up with netem.
	local iface
	for iface in $(tc qdisc show | awk '$2 == "netem" {print $5}' | sort -u); do
		tc qdisc del dev "$iface" root
	done

	echo "cleared"
}

case ${1:-} in
block | delay | loss)
	require_root
	cmd=$1
	shift
	"cmd_$cmd" "$@"
	;;
status)
	require_root
	cmd_status
	;;
clear)
	require_root
	clear_all
	;;
-h | --help | "")
	usage
	;;
*)
	die "unknown command: $1"
	;;
esac
