#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

gofmt -w ./cmd
gofmt -w ./internal
gofmt -w ./main.go
