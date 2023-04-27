#!/bin/bash

docker rm -f $(docker ps -aq)
docker network prune
docker volume prune
docker rmi $(docker images | grep "dev-" | awk '{print $1}') -f
rm -rf ~/.ipfs

