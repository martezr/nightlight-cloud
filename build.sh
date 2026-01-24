# build webui
cd webui
npm run build
cd ..

# build nightlight binary
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nightlight-cloud
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
docker build --platform linux/amd64 -t nightlight-cloud:latest .
docker run --platform linux/amd64 -v $(pwd)/iso:/iso nightlight-cloud:latest
cd ..

# run citesting
echo "Running CI Tests..."
cd citesting
go run main.go