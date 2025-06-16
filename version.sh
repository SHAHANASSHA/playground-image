#!/bin/bash

set -e

git fetch --tags

LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0")

VERSION=${LATEST_TAG#v}
MAJOR=$(echo $VERSION | cut -d. -f1)
MINOR=$(echo $VERSION | cut -d. -f2)

COMMIT_MSG=$(git log -1 --pretty=%B)

if echo "$COMMIT_MSG" | grep -q "#major"; then
    MAJOR=$((MAJOR + 1))
    MINOR=0
else
    MINOR=$((MINOR + 1))
fi

NEW_TAG="v${MAJOR}.${MINOR}"

echo "New tag: $NEW_TAG"

git tag "$NEW_TAG"
git push origin "$NEW_TAG"

echo "$NEW_TAG"

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"
git add "$VERSION_FILE"
if 
git diff --cached --quiet; then
        echo "Nothing to commit"
else
        git commit -m "Version bump to $NEW_VERSION""" [skip ci]"

fi

