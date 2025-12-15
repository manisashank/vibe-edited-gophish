#!/bin/bash

# Build Frontend assets
npm install
npx gulp build

# Build Linux amd64 binary
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o gophish

# Package the release (include db migrations to avoid startup errors)
rm -rf gophish-release-linux64
mkdir -p gophish-release-linux64
cp -r gophish config.json VERSION static templates db gophish-release-linux64/
( cd gophish-release-linux64 && zip -r ../gophish-release-linux64.zip . )

echo "[+] Done"
