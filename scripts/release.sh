#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

CURRENT=$(cat ./internal/gorson/version/version.go | grep 'const Version' | cut -d'"' -f2)
echo $CURRENT
NEW=`expr $CURRENT + 1`

sed -i -e "s/-$CURRENT-/-$NEW-/g" README.md