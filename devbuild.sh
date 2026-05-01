# build webui
cd webui
npm run build
cd ..

env GOOS=linux GOARCH=amd64 go build -o nightlight-cloud
chmod +x nightlight-cloud
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null nightlight-cloud root@10.0.0.237:/tmp

# build metadata agent
cd metadataagent
env GOOS=linux GOARCH=amd64 go build -o metadataagent
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null metadataagent root@10.0.0.237:/tmp

# build dhcp agent
cd dhcpagent
env GOOS=linux GOARCH=amd64 go build -o dhcpagent
chmod +x dhcpagent
scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null dhcpagent root@10.0.0.237:/tmp