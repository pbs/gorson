#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# tooling logic borrowed heavily from the talented minds of confd
# https://github.com/kelseyhightower/confd/blob/master/Makefile

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

./scripts/clean.sh
mkdir -p bin

VERSION=`egrep -o '[0-9]' ./internal/gorson/version/version.go`

>&2 echo "building gorson_builder docker image"
# build a local docker image called gorson_builder: we'll use this to
# compile our release binaries
docker build -t gorson_builder -f Dockerfile.build.alpine .

>&2 echo "compiling binaries for release"
# for each of our target platforms we use the gorson_builder
#   docker container to compile a binary of our application
for platform in darwin linux; do \
    binary_name="gorson-${VERSION}-${platform}-amd64"
    >&2 echo "compiling $binary_name"

    # * GOOS is the target operating system
    # * GOARCH is the target processor architecture
    #     (we only compile for amd64 systems)
    #     see https://golang.org/cmd/go/#hdr-Environment_variables
    # * CGO_ENABLED controls whether the go compiler allows us to
    #     import C packages (we don't do this, so we set it to 0 to turn CGO off)
    #     see https://golang.org/cmd/cgo/
    docker run --rm \
    -v ${PWD}:/app \
    -e "GOOS=$platform" \
    -e "GOARCH=amd64" \
    -e "CGO_ENABLED=0" \
    gorson_builder \
        go build -o bin/$binary_name
done
