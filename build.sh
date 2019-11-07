#!/bin/bash


set -e
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker build -t $DOCKER_ORG/install-controller:$TRAVIS_TAG -t $DOCKER_ORG/install-controller:latest .
docker push $DOCKER_ORG/appsody-controller