#!/bin/bash

VERSION_FILE='version.txt'

if [ ! -f  $VERSION_FILE  ]; then
        echo "0.0" > "$VERSION_FILE"
fi


VERSION=$(read -r VERSION < "$VERSION_FILE")
MAJOR=$(cat $VERSION_FILE | cut -d. -f1)
MINOR=$(cat $VERSION_FILE | cut -d. -f2)

COMMIT_MSG=$(git log -1 --pretty=%B)

if  echo "$VERSION_FILE" | grep -q "#major"; then
        MAJOR=$(( MAJOR + 1 ))
        MINOR=0
else
        MINOR=$(( MINOR + 1 ))
fi

NEW_VERSION="${MAJOR}.${MINOR}"
echo "New version: $NEW_VERSION"
echo "$NEW_VERSION > $VERSION_FILE"

git config user.name "jenkins"
git config user.email "jenkins@j=example.com"
git add "$VERSION_FILE"
git commit -m "Version bump to $NEW_VERSION"
git push origin main

