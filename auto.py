import requests
import time

# Create VPC
vpcPayload = {}
vpcPayload["name"] = "autovpc"
vpcPayload["description"] = "automation vpc"
vpcPayload["cidrBlock"] = "10.0.0.0/16"
print(vpcPayload)
r = requests.post('http://10.0.0.235:3000/vpcs', json=vpcPayload)
vpcResponse = r.json()
print(vpcResponse["id"])

# Create Subnet #1
subnetPayload = {}
subnetPayload["name"] = "autosub01"
subnetPayload["description"] = "auto subnet 01"
subnetPayload["cidrBlock"] = "10.0.0.0/24"
subnetPayload["vpcId"] = vpcResponse["id"]
r = requests.post('http://10.0.0.235:3000/subnets', json=subnetPayload)
print("Request body:", r.request.body)
response = r.json()
print(response["id"])

# Subnet 2
subnet2Payload = {}
subnet2Payload["name"] = "autosub02"
subnet2Payload["description"] = "auto subnet 02"
subnet2Payload["cidrBlock"] = "10.0.1.0/24"
subnet2Payload["vpcId"] = vpcResponse["id"]
r = requests.post('http://10.0.0.235:3000/subnets', json=subnet2Payload)
subnet2response = r.json()
print(subnet2response["id"])

# Image
imagePayload = {}
imagePayload["description"] = "base image"
imagePayload["location"] = "/zfs-pool/images/cirros-0.6.2-x86_64-disk.img"
r = requests.post('http://10.0.0.235:3000/images', json=imagePayload)
imageresponse = r.json()
print(imageresponse["id"])

# Create Instance
#instancePayload = {}
#instancePayload["subnetId"] = subnet2response["id"]
#instancePayload["instanceType"] = "t4.small"
#instancePayload["imageId"] = imageresponse["id"]
#print(instancePayload)
#r = requests.post('http://10.0.0.3:3000/instances', json=instancePayload)
#instanceResponse = r.json()
#print(instanceResponse)