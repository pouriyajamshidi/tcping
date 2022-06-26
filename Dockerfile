FROM scratch

ADD executables/tcping /usr/bin/tcping

ENTRYPOINT ["tcping"]
