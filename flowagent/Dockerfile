FROM golang:1.25

# Install eBPF build dependencies
RUN apt-get update && apt-get install -y \
    clang \
    llvm \
    libbpf-dev \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

# Build eBPF program
RUN clang -O2 -g -target bpf -c flow.c -o flow.o

# Build Go binary
RUN mkdir -p /output
CMD [ "sh", "-c", "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /output/flowmonitoragent" ]