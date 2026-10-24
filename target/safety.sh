#!/usr/bin/env bash
# safety.sh — runs the pre-detonation "safety check" with the supplied code.
# In the real exercise this verifies the device is in position. Here it is a
# harmless stub: it records the check and exits 0. It does NOTHING destructive.
#
# NOTE: trigger.go invokes this as:  bash -c "./safety.sh <detonationCode>"
# so <detonationCode> is attacker-controlled and unsanitized. That is the bug.

CODE="$*"
STATEDIR="${STATEDIR:-$(dirname "$0")}"
echo "[safety] $(date '+%H:%M:%S') running safety check with code: ${CODE}" >> "${STATEDIR}/bomb.log"
echo "safety check OK (code=${CODE})"
exit 0
