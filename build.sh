# build webui
cd webui
npm run build
cd ..

# build nightlight binary
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nightlight-cloud
chmod +x nightlight-cloud
cp nightlight-cloud isobuild/

# build appliance config binary
echo "Building Config Binary..."
cd applianceconfig
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o nightlight-config
chmod +x nightlight-config
cp nightlight-config ../isobuild/
cd ..

# # build metadata agent
# echo "Building Metadata Agent..."
# cd metadataagent
# env GOOS=linux GOARCH=amd64 go build -o metadataagent
# chmod +x metadataagent
# cp metadataagent ../isobuild/
# cd ..

# build flow monitor agent
echo "Building Flow Monitor Agent..."
cd flowagent
env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o flowmonitoragent
chmod +x flowmonitoragent
cp flowmonitoragent ../isobuild/
cd ..

# # build dhcp agent
# echo "Building DHCP Agent..."
# cd dhcpagent
# env GOOS=linux GOARCH=amd64 go build -o dhcpagent
# chmod +x dhcpagent
# cp dhcpagent ../isobuild/
# cd ..

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