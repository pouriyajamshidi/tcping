# A probe through a proxy needs a proxy, so this tape starts tinyproxy in a
# container off camera and removes it at the end. Needs docker on the PATH,
# and network access the first time, to pull the image and install tinyproxy.
#
# The size is the same as tcping_http_verbose.tape, which runs the same probe
# without the proxy: its widest line, a probe line carrying an IPv6 address,
# is wider than the start line naming the proxy.

Set Shell "bash"
Set FontSize 20
Set Padding 30
Set Width 1475
Set Height 1015

Hide
Type "docker run -d --rm --name tcping-gif-proxy -p 127.0.0.1:3128:8888 alpine:3.24 sh -c 'apk add --no-cache tinyproxy && sed -i s/^Allow/#Allow/ /etc/tinyproxy/tinyproxy.conf && tinyproxy -d' >/dev/null"  Enter
Type "until docker logs tcping-gif-proxy 2>&1 | grep -q Accepting; do sleep 0.5; done; printf 'proxy %s\n' up"  Enter
Wait+Screen@120s /(?m)^proxy up$/
Type "clear"  Enter
Sleep 500ms
Show

Type "tcping https://example.com --proxy http://127.0.0.1:3128 -v -c 3"
Sleep 500ms  Enter

Sleep 15s

Hide
Type "docker rm -f tcping-gif-proxy >/dev/null"  Enter
Sleep 2s
