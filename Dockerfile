FROM golang:1.24 AS builder
WORKDIR /build
ENV PACKAGE="github.com/netresearch/udp-proxy/internal/build"
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . /build
RUN export VERSION="$(git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')"
RUN export COMMIT="$(git rev-parse --short HEAD)"
RUN export BUILD_TIME=$(date '+%Y-%m-%dT%H:%M:%S')
RUN go build -o /udp-proxy -ldflags="-s -w -X '${PACKAGE}.Version=${VERSION}' -X '${PACKAGE}.Commit=${COMMIT}' -X '${PACKAGE}.BuildTime=${BUILD_TIME}'"

FROM alpine:3.21 AS runner
WORKDIR /data

COPY --from=builder /udp-proxy /bin/udp-proxy
ENTRYPOINT [ "/bin/udp-proxy" ]
