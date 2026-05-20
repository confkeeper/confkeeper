#!/bin/bash
set -e

# Default values
WEB_REPO=confkeeper/confkeeper-web
WEB_VERSION="v1.1.1"

if [ -z "$RELEASE_TOKEN" ]; then
  echo "Error: RELEASE_TOKEN must be set"
  exit 1
fi

echo "Downloading web dist from $WEB_REPO ($WEB_VERSION)..."

# Ensure GITHUB_TOKEN is available for gh cli
export GITHUB_TOKEN="$RELEASE_TOKEN"

# Use gh cli to download the asset, which handles auth and redirects better than curl
gh release download "$WEB_VERSION" \
  --repo "$WEB_REPO" \
  --pattern "dist.zip" \
  --dir "." \
  --clobber

echo "Extracting dist.zip..."
mkdir -p static/
unzip -q -o dist.zip -d static/
rm dist.zip

echo "Web dist downloaded and extracted to static/"
