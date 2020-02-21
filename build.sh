#!/bin/bash


set -e
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker buildx build -t $DOCKER_ORG/init-controller:$TRAVIS_TAG -t $DOCKER_ORG/init-controller:latest --platform=linux/amd64,linux/ppc64le . --push