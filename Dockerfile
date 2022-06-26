FROM scratch

ADD tcpingDocker /usr/bin/tcping

ENTRYPOINT ["tcping"]
