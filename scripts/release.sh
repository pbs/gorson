#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

function is_gnu_sed() {
  sed --version >/dev/null 2>&1
}

function find_and_replace() {
    EXPRESSION="$1"
    FILE="$2"
    if is_gnu_sed; then
        sed -i -r "$EXPRESSION" "$FILE"
    else
        sed -i '' -E "$EXPRESSION" "$FILE"
    fi
}

cd "$( dirname "${BASH_SOURCE[0]}" )/.."

CURRENT=$(grep -E -o '[0-9]' ./internal/gorson/version/version.go)
NEW=$(( CURRENT + 1))

# Internal gorson version
find_and_replace "s/$CURRENT/$NEW/g" ./internal/gorson/version/version.go
# Binary version
find_and_replace "s/gorson-$CURRENT-([^-]+)-amd64/gorson-$NEW-\1-amd64/g" README.md
# asdf version
find_and_replace "s/gorson $CURRENT/gorson $NEW/g" README.md

git checkout -b "release/$NEW"
git add README.md ./internal/gorson/version/version.go
git commit -m "release $NEW"
git tag $NEW
git push -u origin "release/$NEW"
git push --tags
