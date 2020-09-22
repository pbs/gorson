#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

CURRENT=$(egrep -o '[0-9]' ./internal/gorson/version/version.go)
NEW=`expr $CURRENT + 1`

sed "-i" "" "-e" "s/-$CURRENT-/-$NEW-/g" README.md
sed "-i" "" "-e" "s/$CURRENT/$NEW/g" ./internal/gorson/version/version.go

git checkout -b "release/$NEW"
git add README.md ./internal/gorson/version/version.go
git commit -m "release $NEW"
git tag $NEW
git push -u origin "release/$NEW"
git push --tags
