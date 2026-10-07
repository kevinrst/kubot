# pod_crashloop_backoff

A container keeps crashing and kubelet keeps restarting it.

Severity: critical

## What kubot saw

The container is in `CrashLoopBackOff`, or it has 3+ restarts with a recent
non-zero exit (the backoff window between crashes looks like `Error` or a
fresh `Terminated` state — same loop, different snapshot instant). Only
crashes from the last 15 minutes count; old restarts on a healthy container
are history, not a diagnosis. OOMKills are excluded here — they get their own
finding with memory evidence.

## Why it matters

The workload is down or flapping. Every restart is downtime, and backoff
delays mean recovery gets slower the longer it loops.

## Verify it yourself

```sh
kubectl logs <pod> --previous   # the crash output, not the current attempt
kubectl describe pod <pod>      # last termination: exit code and reason
```

## Fix

Read the previous-container logs first — the answer is usually there (bad
command, missing env, failed migration). Then check the startup path, mounted
config, and probes.

## When to ignore

Rarely — a crashloop is never "tolerated". If a job-like pod is *supposed* to
exit nonzero (it isn't; use a Job), mute it:

```toml
[[ignore]]
finding = "pod_crashloop_backoff"
object = "default/batch-*"
reason = "exits nonzero by design (should be a Job)"
```

## What kubot can't see

Application logs beyond the exit code. kubot tells you *that* it crashes and
*how* (exit code, restarts); the *why inside the app* is in the logs.
