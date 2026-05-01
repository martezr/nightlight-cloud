# build the docker image
docker build --platform linux/amd64 -t flowmonitoragent:latest .
docker run --platform linux/amd64 -v $(pwd):/output flowmonitoragent:latest