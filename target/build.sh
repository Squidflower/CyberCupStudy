#!/usr/bin/env bash
# Build the wmd-trigger target binary. Needs Go 1.20+ (https://go.dev/dl/).
set -e
cd "$(dirname "$0")"
chmod +x safety.sh detonate
go build -o wmd-trigger trigger.go
echo "[+] built ./target/wmd-trigger"
