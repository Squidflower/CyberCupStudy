#!/usr/bin/env bash
set -e
HAIR="false"; SIG="false"
case "$LEVEL" in
  warmup)   : ;;
  standard) HAIR="true" ;;
  hard)     HAIR="true"; SIG="true" ;;
  random)   [ $((RANDOM%2)) -eq 0 ] && HAIR="true"; [ $((RANDOM%2)) -eq 0 ] && SIG="true" ;;
esac
ARGS=(-listenPort "$PORT")
[ "$HAIR" = "true" ] && ARGS+=(-hairTrigger)
[ "$SIG" = "true" ]  && ARGS+=(-catchSignals)
# config written inside the container only; `docker exec <c> cat /opt/trigger/.config` to check after
printf 'level=%s hairTrigger=%s catchSignals=%s minutes=%s\n' "$LEVEL" "$HAIR" "$SIG" "$MINUTES" > /opt/trigger/.config
exec env MINUTES_TILL_SELF_DESTRUCT="$MINUTES" ./wmd-trigger "${ARGS[@]}"
