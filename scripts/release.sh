#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

CURRENT=$(cat ./internal/gorson/version/version.go | grep 'const Version' | cut -d'"' -f2)
NEW=`expr $CURRENT + 1`

sed "-i" "" "-e" "s/-$CURRENT-/-$NEW-/g" README.md
sed "-i" "" "-e" "s/$CURRENT/$NEW/g" ./internal/gorson/version/version.go

git checkout -b "release/$NEW"
git add README.md ./internal/gorson/version/version.go
git commit -m "release $NEW"
git tag $NEW
git push
git push --tags
