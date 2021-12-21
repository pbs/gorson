#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

>&2 echo "running gorson tests"
go test -v -count=1 ./...
