# Build stage
##################################################
FROM docker.io/golang:1.26.7-alpine3.23 AS build

WORKDIR /build

# Install dependencies
# git is needed for the Makefile to derive VERSION from git tags/branch
RUN apk --no-cache add bash make git

# Cache libraries
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY ./ ./
RUN make build

# Final stage
##################################################
FROM scratch

LABEL maintainer="Pouriya Jamshidi"

COPY --from=build /build/target/tcping /usr/bin/

ENTRYPOINT ["tcping"]
