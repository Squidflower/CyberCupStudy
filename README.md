# Fiery Gambit — Practice Range

A self-contained lab that recreates the **Fiery Gambit** exercise so you can rehearse
the full solve against a *real* listening target: recon → command injection → deny the
trigger → disarm before the countdown. Built from the exercise's own `trigger.go`.

Everything here is a **training range**. The `safety.sh` and `detonate` scripts are
harmless stubs — a "detonation" just writes a `DETONATED` marker and prints a banner.
Nothing destructive runs. Use it only on machines you own (your box or your homelab).

---

## What's in the box

```
fiery-gambit-lab/
├─ target/
│  ├─ trigger.go       the device (the exercise's captured source, lightly annotated)
│  ├─ safety.sh        harmless "safety check" stub (the injection sink)
│  ├─ detonate         harmless "detonation" stub (simulated boom)
│  └─ build.sh         compiles ./target/wmd-trigger   (needs Go 1.20+)
├─ deploy-bomb.sh      launch ONE bomb at a chosen difficulty
├─ range.sh            launch several bombs with random configs
├─ stop.sh             kill + clean the whole range
├─ Dockerfile          \
├─ docker-entrypoint.sh } containerized targets with real IPs to nmap
├─ docker-compose.yml  /
├─ mission-order.txt   the brief, for reference
├─ README.md           this file
└─ SOLUTION.md         full walkthrough — don't open until you've tried
```

---

## Quick start (bare metal — simplest)

Needs: `go`, `nc` (netcat), and ideally `nmap`. On Kali you already have them.

```bash
./target/build.sh                 # compile the device
./deploy-bomb.sh alpha 9803 warmup 20
```

Then, from another terminal, run the solve against `127.0.0.1:9803`.
When you finish (or it blows), check how you did:

```bash
cat runs/alpha/.config      # what traps were actually on
cat runs/alpha/bomb.log     # what the device logged
ls  runs/alpha/DETONATED    # present ONLY if it detonated
./stop.sh                   # clean up
```

Difficulty levels for `deploy-bomb.sh <id> <port> <level> [minutes]`:

| level      | hairTrigger | signal trap | what it teaches |
|------------|:-----------:|:-----------:|-----------------|
| `warmup`   |  off  | off | the clean flow: inject, disarm |
| `standard` |  **on** | off | don't just drop your session |
| `hard`     |  **on** | **on** | naive `kill`/Ctrl-C detonates — use `kill -9` |
| `random`   |  ? | ? | unknown config — recon to find out (most like the real thing) |

Launch a whole field of devices with mixed, hidden configs:

```bash
./range.sh 3
nmap -sV -p 9803-9805 127.0.0.1
```

## Docker (real network IPs to recon)

```bash
docker compose up --build -d
nmap -sV -p 9803 172.30.0.11 172.30.0.12 172.30.0.13
# attack each at <ip>:9803
docker compose down
```

---

## The flow you're drilling

1. **Develop capabilities.** Read `target/trigger.go`. Find where your input reaches a shell.
2. **Recon.** Find the device and its open port (`nmap -sV`).
3. **Access.** Reach the port and get code execution through the input.
4. **Local recon.** Identify the trigger process; *account for unknown modifications*
   (is it `hairTrigger`? does it catch signals?).
5. **Deny & disarm.** Stop the trigger process so it can't fire — safely — before the clock.

The traps are the lesson:
- **hairTrigger** detonates when your session drops, so don't rely on just disconnecting.
- A **signal trap** can turn a plain `kill` (SIGTERM) or Ctrl-C (SIGINT) into a detonation,
  so the safe stop is **`kill -9`** (SIGKILL can't be caught).
- A **self-destruct** countdown is always running — you're racing it.

Stuck? `SOLUTION.md` has progressive hints before the full answer.

> Reminder: lab targets only — your own machine or homelab. This recreates a sanctioned
> collegiate exercise for practice; keep it on infrastructure you control.
