#!/bin/bash
set -e
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
export DOCKER_CLI_EXPERIMENTAL=enabled
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker buildx create --name mybuilder
docker buildx use mybuilder
docker buildx build -t appsody/init-controller:$TRAVIS_TAG -t appsody/init-controller:latest --platform=linux/amd64,linux/ppc64le,linux/s390x . --push