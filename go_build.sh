#!/bin/bash
set -e
for OS_ARCH in linux/amd64 linux/ppc64le linux/s390x
do
   declare OS=$(echo $OS_ARCH | cut -f1 -d/)
   declare ARCH=$(echo $OS_ARCH | cut -f2 -d/)
   GOOS="$OS" CGO_ENABLED=0 GOARCH="$ARCH" go build -o ./build/appsody-controller-"$OS_ARCH" -ldflags "-X main.VERSION=$1"
done
