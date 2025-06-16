#!/bin/bash

VERSION_FILE='version.txt' 

if [ ! -f  "$VERSION_FILE" ]; then
        echo "0.0" > "$VERSION_FILE"          #Creates a "version.txt file if not existing and redirect the current version on it."
fi


VERSION=$(cat "$VERSION_FILE")                #Read the .txt file.
MAJOR=$(echo "$VERSION" | cut -d. -f1)        #Cut the first part of the tag befor " . " 
MINOR=$(echo "$VERSION" | cut -d. -f2)        #Cut the second part of the tag after " . "

COMMIT_MSG=$(git log -1 --pretty=%B)          #Pretty the last commit in oneline. 

if  echo "$COMMIT_MSG" | grep -q "#major"; then #Read the commit msg and check for the tag major. if the major exist it will go for major update, otherwise go for minor.
        MAJOR=$((MAJOR + 1))                        
        MINOR=0
else
        MINOR=$((MINOR + 1))
fi

NEW_VERSION="${MAJOR}.${MINOR}"                #Compain thos two values in to a variable.
echo "New version: $NEW_VERSION"               #Print the new version.
echo "$NEW_VERSION" > "$VERSION_FILE"          #Re-direct the value to the variable.

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"
git add "$VERSION_FILE"
if 
git diff --cached --quiet; then
        echo "Nothing to commit"
else
        git commit -m "Version bump to $NEW_VERSION""" [skip ci]"

fi
