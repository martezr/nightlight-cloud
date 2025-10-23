env GOOS=linux GOARCH=amd64 go build -o nightlight-cloud
chmod +x nightlight-cloud
scp nightlight-cloud root@10.0.0.235:/tmp