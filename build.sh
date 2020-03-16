#!/bin/bash
#!/bin/bash


set -e
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
export DOCKER_CLI_EXPERIMENTAL=enabled
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
docker buildx create --name mybuilder
docker buildx use mybuilder
docker buildx build -t $DOCKER_ORG/init-controller:$TRAVIS_TAG -t $DOCKER_ORG/init-controller:latest --platform=linux/amd64,linux/ppc64le,linux/s390x . --push