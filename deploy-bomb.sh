#!/usr/bin/env bash
# deploy-bomb.sh — launch ONE practice bomb.
#
#   ./deploy-bomb.sh <id> <port> <level> [minutes]
#
# level:
#   warmup    no traps. clean injection + kill. learn the flow.
#   standard  hairTrigger ON  (dropping your session detonates)
#   hard      hairTrigger + signal trap ON  (naive kill/Ctrl-C detonates too)
#   random    traps chosen at random — you won't know until you recon. (realistic)
#
# The trap config is hidden in runs/<id>/.config so you don't spoil yourself.
# Check it AFTER an attempt to see what you were up against.
set -e
cd "$(dirname "$0")"

ID="${1:-alpha}"
PORT="${2:-9803}"
LEVEL="${3:-random}"
MINUTES="${4:-20}"

BIN="target/wmd-trigger"
[ -x "$BIN" ] || { echo "[!] build first:  ./target/build.sh"; exit 1; }

HAIR="false"; SIG="false"
case "$LEVEL" in
  warmup)   HAIR="false"; SIG="false" ;;
  standard) HAIR="true";  SIG="false" ;;
  hard)     HAIR="true";  SIG="true"  ;;
  random)   [ $((RANDOM%2)) -eq 0 ] && HAIR="true"; [ $((RANDOM%2)) -eq 0 ] && SIG="true" ;;
  *) echo "[!] unknown level: $LEVEL"; exit 1 ;;
esac

STATE="$(pwd)/runs/$ID"          # absolute, so paths survive the cd into target/
TARGETDIR="$(pwd)/target"
mkdir -p "$STATE"; rm -f "$STATE/DETONATED" "$STATE/bomb.log"
printf 'level=%s hairTrigger=%s catchSignals=%s minutes=%s port=%s\n' \
  "$LEVEL" "$HAIR" "$SIG" "$MINUTES" "$PORT" > "$STATE/.config"

ARGS=(-listenPort "$PORT")
[ "$HAIR" = "true" ] && ARGS+=(-hairTrigger)
[ "$SIG" = "true" ]  && ARGS+=(-catchSignals)

# run with cwd=target/ so ./safety.sh and ./detonate resolve; per-bomb STATEDIR.
# exec replaces the subshell so $! is the real wmd-trigger PID; detach from this terminal.
( cd "$TARGETDIR" && STATEDIR="$STATE" BOMB_ID="$ID" MINUTES_TILL_SELF_DESTRUCT="$MINUTES" \
    exec ./wmd-trigger "${ARGS[@]}" >"$STATE/trigger.out" 2>&1 <&- ) &
echo $! > "$STATE/pid"
disown 2>/dev/null || true
sleep 0.4
PID="$(cat "$STATE/pid")"

cat <<EOF
############################################################
  DEVICE DEPLOYED  —  bomb "$ID"
  target : 127.0.0.1 : $PORT      (or this host's LAN IP)
  armed  : self-destruct in $MINUTES minute(s)
  level  : $LEVEL   (trap config hidden — recon to find out)
############################################################
  Your job: deny the trigger and disarm before it blows.
  When done (or boom), check:
    cat $STATE/.config      # what traps were really on
    cat $STATE/bomb.log     # what the device logged
    ls  $STATE/DETONATED    # exists only if it went off
  Stop/clean everything:  ./stop.sh
EOF
