# Build stage
##################################################
FROM docker.io/golang:1.23-alpine3.20 AS build

WORKDIR /build

# Install dependencies
RUN apk --no-cache add bash make

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
