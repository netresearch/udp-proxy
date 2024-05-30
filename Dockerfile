FROM golang:1.22 as builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . /build
RUN CGO_ENABLED=0 go build -o /udp-proxy

FROM alpine:3.20 AS runner
WORKDIR /data

COPY --from=builder /udp-proxy /bin/udp-proxy
ENTRYPOINT [ "/bin/udp-proxy" ]