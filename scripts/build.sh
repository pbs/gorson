#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# tooling logic borrowed heavily from the talented minds of confd
# https://github.com/kelseyhightower/confd/blob/master/Makefile

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

>&2 echo "building gorson"
mkdir -p bin
go build -o bin/gorson .
