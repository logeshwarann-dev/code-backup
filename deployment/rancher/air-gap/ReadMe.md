#!/bin/bash

#https://hangar.cnrancher.com/docs/v1.7/bestpractice/rke2/
# In this example, we create a directory to store the container image layers.
mkdir -p registry

docker run -d \
    -p 5000:5000 \
    -v $(pwd)/registry:/var/lib/registry \
    --name registry \
    registry:2

hangar login 'localhost:5000' --tls-verify=false


hangar load \
    -s 'rke2-images.zip' \
    -d 'localhost:5000' \
    --arch 'amd64' \
    --os 'linux' \
    --jobs 5 \
    --tls-verify=false