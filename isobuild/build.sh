#!/bin/bash
# build the golang binary for linux

CGO_ENABLED=0 GOOS=linux go build -o nightlight-cloud
chmod +x nightlight-cloud

# build the docker image
docker build -t nightlight-cloud:latest .
docker run -v $(pwd)/iso:/iso nightlight-cloud:latest