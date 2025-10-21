# Nightlight Cloud

Nightlight cloud is a lightweight virtualization solution that is intended for homelab environments and lacks many of the standard enterprise grade virtualization features like clustering, live migration, and more. The goal is to enable rapid expirementation 

curl -X POST http://10.0.0.102:3000/vpcs -d '{"name":"demovpc01","description":"demo vpc","cidrBlock":"10.0.0.0/16"}'

curl -X POST http://10.0.0.235:3000/instances -d '{"subnetId":"subnet-3h1xf4maej","instanceType":"t4.small","imageId":"image-xqymkhs80k","datastoreId":"datastore-3f9qklopi1"}'