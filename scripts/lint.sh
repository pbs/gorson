#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

>&2 echo 'linting'
invalid=$(gofmt -l . | grep -v vendor 2>&1)
echo "a"
if [ "$invalid" ]; then
  echo "b"
  >&2 echo "linting failed on the following files; ./scripts/format.sh should fix this"
  >&2 echo $invalid
  exit 1
fi