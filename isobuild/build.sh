#!/bin/bash

tar -C overlay -czf nightlight.apkovl.tar.gz .
cp nightlight.apkovl.tar.gz overlaytar

# build the docker image
docker build --platform linux/amd64 -t nightlight-cloud:latest .
docker run --platform linux/amd64 -v $(pwd)/iso:/iso -v $PWD/overlaytar:/overlay nightlight-cloud:latest