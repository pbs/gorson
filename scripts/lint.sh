#!/usr/bin/env bash
set -uo pipefail
IFS=$'\n\t'

# I'd love a simpler way to do this check, but no native way seems to exist
# https://github.com/golang/go/issues/24230
# https://github.com/golang/go/issues/24427


>&2 echo 'linting'
invalid="$(gofmt -s -l . | grep -v vendor)"
test -z "$invalid"
if [[ $? -ne 0 ]]; then
  >&2 echo "linting failed on the following files; ./scripts/format.sh should fix this"
  >&2 echo $invalid
  exit 1
fi
echo 'linting succeeded :)'