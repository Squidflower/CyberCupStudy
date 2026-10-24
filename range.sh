#!/usr/bin/env bash
# range.sh — deploy MULTIPLE devices at once (the "one or more devices" scenario).
# Each gets a random level, so you must recon and analyze each one.
#
#   ./range.sh [count]        default 3
set -e
cd "$(dirname "$0")"

COUNT="${1:-3}"
BASEPORT=9803
NAMES=(alpha bravo charlie delta echo foxtrot golf hotel)

echo "[*] deploying $COUNT device(s)..."
for i in $(seq 0 $((COUNT-1))); do
  ID="${NAMES[$i]:-bomb$i}"
  PORT=$((BASEPORT+i))
  MIN=$(( (RANDOM%3)*5 + 10 ))     # 10, 15, or 20 minutes
  ./deploy-bomb.sh "$ID" "$PORT" random "$MIN" | sed -n '2,6p'
  echo
done

echo "============================================================"
echo "  TARGET LIST"
for i in $(seq 0 $((COUNT-1))); do
  ID="${NAMES[$i]:-bomb$i}"; PORT=$((BASEPORT+i))
  echo "    $ID   ->  127.0.0.1:$PORT"
done
echo "  Scan them:   nmap -sV -p $BASEPORT-$((BASEPORT+COUNT-1)) 127.0.0.1"
echo "  Stop all:    ./stop.sh"
echo "============================================================"
