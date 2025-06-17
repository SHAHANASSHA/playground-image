#!/bin/bash

set -e

# Get latest tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Last tag: $LAST_TAG"

# Get commits since last tag
COMMITS=$(git log ${LAST_TAG}..HEAD --pretty=format:"%s%n%b")

echo "Commits since $LAST_TAG:"
echo "$COMMITS"

BUMP="patch"

if echo "$COMMITS" | grep -q "major"; then
  BUMP="major"
elif echo "$COMMITS" | grep -q "^minor:"; then
  BUMP="minor"
elif echo "$COMMITS" | grep -q "^patch:"; then
  BUMP="patch"
else
  echo "No version bump needed. Exiting."
  exit 0
fi

# Parse version parts
VERSION=${LAST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Bump
case "$BUMP" in
  major)
    ((MAJOR+=1)); MINOR=0; PATCH=0 ;;
  minor)
    ((MINOR+=1)); PATCH=0 ;;
  patch)
    ((PATCH+=1)) ;;
esac

NEW_TAG="v$MAJOR.$MINOR.$PATCH"
echo "Creating tag $NEW_TAG"

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"

git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

