echo "Building nightlight-cloud binary..."
env GOOS=linux GOARCH=amd64 go build -o nightlight-cloud
chmod +x nightlight-cloud
cp nightlight-cloud isobuild/

# build metadata agent
echo "Building Metadata Agent..."
cd metadataagent
env GOOS=linux GOARCH=amd64 go build -o ../metadataagent
cd ..

# build the docker image
echo "Building ISO Docker Image..."
cd isobuild
docker build -t nightlight-cloud:latest .
docker run -v $(pwd)/iso:/iso nightlight-cloud:latest
cd ..

# run citesting
echo "Running CI Tests..."
cd citesting
go run main.go