# pod_oom_killed

A container was killed for exceeding its memory limit (exit 137).

Severity: critical

## What kubot saw

A current or previous termination with reason `OOMKilled`, plus the
container's memory limit and restart count.

## Why it matters

The process doesn't die from a bug — it dies from arithmetic. It will keep
dying on every restart until the limit or the usage changes. No amount of
replicas fixes a per-container limit.

## Verify it yourself

```sh
kubectl describe pod <pod>      # Last State: OOMKilled, exit 137
kubectl top pod <pod>           # live usage vs the limit (needs metrics-server)
```

## Fix

Either raise the limit or cut usage. If usage climbs steadily before death,
it's a leak — profile the app. If it spikes, the limit is just too small.

## When to ignore

Almost never. The one case: a workload that legitimately bursts past its
limit and is *allowed* to be OOMKilled and restarted (rare, and usually a
design smell):

```toml
[[ignore]]
finding = "pod_oom_killed"
object = "default/worker-*"
reason = "burst workload, restarts are the design"
```

## What kubot can't see

Whether the memory growth is a leak or legitimate load — that needs profiling,
not cluster state.
