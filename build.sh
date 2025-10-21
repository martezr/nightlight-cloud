env GOOS=linux GOARCH=amd64 go build -o nightlight-cloud
chmod +x nightlight-cloud
cp nightlight-cloud isobuild/

# build the docker image
cd isobuild
docker build -t nightlight-cloud:latest .
docker run -v $(pwd)/iso:/iso nightlight-cloud:latest
cd ..

# run citesting
cd citesting
go run main.go