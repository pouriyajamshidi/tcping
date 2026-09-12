# Build stage
##################################################
# Pinned to the build machine's own platform so the compiler always runs
# natively. Go cross-compiles to TARGETARCH below, so there is no emulation.
FROM --platform=$BUILDPLATFORM docker.io/golang:1.27.1-alpine3.23 AS build

WORKDIR /build

# Install dependencies
RUN apk --no-cache add bash make

# Cache libraries
COPY go.mod go.sum ./
RUN go mod download

# Build
# Set by buildx, one value per platform in the build. Empty for a plain
# "docker build", where an unset GOARCH means the host architecture anyway.
ARG TARGETARCH
COPY ./ ./
RUN GOARCH=$TARGETARCH make build

# Final stage
##################################################
FROM scratch

LABEL maintainer="Pouriya Jamshidi"

COPY --from=build /build/target/tcping /usr/bin/

ENTRYPOINT ["tcping"]
