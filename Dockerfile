FROM scratch

LABEL maintainer="Pouriya Jamshidi"

COPY tcpingDocker /usr/bin/tcping

ENTRYPOINT ["tcping"]
