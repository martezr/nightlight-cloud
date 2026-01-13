#!/bin/bash
# build the golang binary for linux

CGO_ENABLED=0 GOOS=linux go build -o nightlight-cloud
chmod +x nightlight-cloud

# build the docker image
docker build --platform linux/amd64 -t nightlight-cloud:latest .
docker run --platform linux/amd64 -v $(pwd)/iso:/iso nightlight-cloud:latest