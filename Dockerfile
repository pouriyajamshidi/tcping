FROM scratch
ARG GOOS

ADD execuatables/tcping_${GOOS} /usr/bin/tcping

ENTRYPOINT [ "tcping" ]