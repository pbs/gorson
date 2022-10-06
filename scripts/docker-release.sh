#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# tooling logic borrowed heavily from the talented minds of confd
# https://github.com/kelseyhightower/confd/blob/master/Makefile

cd "$(dirname "${BASH_SOURCE[0]}")/.."

./scripts/clean.sh
mkdir -p bin

# We want to make sure the final builds are formatted and linted properly.
./scripts/format.sh
./scripts/lint.sh

# for each of our target platforms we use the gorson_builder
#   docker container to compile a binary of our application
echo >&2 "compiling binaries for release"
for architecture in amd64 arm64; do
    for platform in darwin linux; do
        binary_name="gorson-${platform}-${architecture}"
        echo >&2 "compiling $binary_name"

        # * GOOS is the target operating system
        # * GOARCH is the target processor architecture
        #     see https://golang.org/cmd/go/#hdr-Environment_variables
        # * CGO_ENABLED controls whether the go compiler allows us to
        #     import C packages (we don't do this, so we set it to 0 to turn CGO off)
        #     see https://golang.org/cmd/cgo/
        GOOS=$platform GOARCH=$architecture CGO_ENABLED=0 go build -o "bin/$binary_name"
    done
done
