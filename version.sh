#!/bin/bash

LATEST_TAG=$(git describe --tags --abbrev=0 >/dev/null || echo "v0.0") 
LATEST_TAG=${LATEST_TAG#v}
MAJOR=$(echo "$LATEST_TAG" | cut -d. -f1)       
MINOR=$(echo "$LATEST_TAG" | cut -d. -f2)       

COMMIT_MSG=$(git log -1 --pretty=%B)         
if  echo "$COMMIT_MSG" | grep -q "#major"; then 
        MAJOR=$((MAJOR + 1))                        
        MINOR=0
else
        MINOR=$((MINOR + 1))
fi

NEW_VERSION="${MAJOR}.${MINOR}"            
echo "$NEW_VERSION"                   

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"

