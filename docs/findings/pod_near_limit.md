# pod_near_limit

A container using 80%+ of its memory limit — a future OOM, still preventable.

Severity: warning

## What kubot saw

Live usage from metrics-server against the configured limit, as a ratio.
Needs metrics-server; without it this rule sleeps and says so (degraded),
instead of guessing.

## Why it matters

This is the finding that precedes `pod_oom_killed`. Memory doesn't degrade
gracefully — at 100% the kernel kills, no throttle, no warning. 80% is the
last moment action is cheap.

## Verify it yourself

```sh
kubectl top pod <pod>   # live usage (needs metrics-server)
```

## Fix

Raise the limit or cut usage now, before the next spike decides for you.

## When to ignore

Workloads that legitimately sit hot (caches sized to fill memory):

```toml
[[ignore]]
finding = "pod_near_limit"
object = "default/redis-*"
reason = "cache sized to use its box"
```

## What kubot can't see

The future — usage *rate of change* would say how urgent 80% is, and kubot
has no history yet. Treat steady-hot and climbing-hot the same for now.
