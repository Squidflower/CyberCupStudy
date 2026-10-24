#!/usr/bin/env bash
# stop.sh — kill every running practice bomb and clean the range.
cd "$(dirname "$0")"
pkill -9 -x wmd-trigger 2>/dev/null && echo "[+] stopped all wmd-trigger processes" || echo "[*] none running"
rm -rf runs/
echo "[+] cleared runs/ (logs, markers, pids)"
