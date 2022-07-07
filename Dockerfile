FROM scratch

COPY tcpingDocker /usr/bin/tcping

ENTRYPOINT ["tcping"]
