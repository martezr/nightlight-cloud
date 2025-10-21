import requests
import time

# Create VPC
r = requests.get('http://10.0.0.3:3000/vpcs')
vpcResponse = r.json()
for vpc in vpcResponse["vpcs"]:
    url = "http://10.0.0.3:3000/vpcs/" + vpc["id"]
    demo = requests.delete(url)
    print(demo)

# Cleanup Subnets
r = requests.get('http://10.0.0.3:3000/subnets')
subnetResponse = r.json()
for subnet in subnetResponse["subnets"]:
    url = "http://10.0.0.3:3000/subnets/" + subnet["id"]
    demo = requests.delete(url)
    print(demo)