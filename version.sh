#!/bin/bash

LATEST_TAG=$(gitdiscribe --tags --abbrev=0 >/dev/null || echo "0.0") 
LATEST_TAG=${LATEST_TAG#v}
MAJOR=$(echo "$VERSION" | cut -d. -f1)       
MINOR=$(echo "$VERSION" | cut -d. -f2)       

COMMIT_MSG=$(git log -1 --pretty=%B)         
if  echo "$COMMIT_MSG" | grep -q "#major"; then 
        MAJOR=$((MAJOR + 1))                        
        MINOR=0
else
        MINOR=$((MINOR + 1))
fi

NEW_VERSION="${MAJOR}.${MINOR}"            
echo "New version: $NEW_VERSION"                   

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"
git add "$VERSION_FILE"
if 
git diff --cached --quiet; then
        echo "Nothing to commit"
else
        git commit -m "Version bump to $NEW_VERSION"

fi
