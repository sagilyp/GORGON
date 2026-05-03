FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN apt-get update && apt-get install -y \
    clang \
    llvm \
    libbpf-dev \
    gcc-multilib \
    linux-libc-dev
RUN go generate ./filter/ebpf/...
RUN go build -o gorgon ./main.go

FROM ubuntu:22.04
WORKDIR /app
RUN apt-get update && apt-get install -y iproute2 iproute2-doc libbpf0 ca-certificates curl jq && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/gorgon /app/gorgon
COPY --from=builder /app/ips /app/ips
RUN mkdir -p /var/log/gorgon /var/lib/tor
CMD ["/app/gorgon"]