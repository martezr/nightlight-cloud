#!/bin/bash

env GOOS=linux GOARCH=amd64 go build -o tui
chmod +x tui
scp -o "StrictHostKeyChecking=no" -o "UserKnownHostsFile=/dev/null" tui root@10.0.0.237:/tmp