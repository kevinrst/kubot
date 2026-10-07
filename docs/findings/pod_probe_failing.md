# pod_probe_failing

A readiness, liveness, or startup probe keeps failing.

Severity: warning

## What kubot saw

Repeated `Unhealthy` warning events naming the probe (3+ aggregated
failures — one-off startup blips stay silent), on a pod that is currently
*not* Ready (a Ready pod's history is stale by definition).

## Why it matters

Readiness gates traffic: a failing readiness probe means the pod gets no
requests even though it runs. Liveness kills and restarts the container.
Startup blocks everything behind it.

## Verify it yourself

```sh
kubectl describe pod <pod>   # Unhealthy events with the probe detail
kubectl port-forward <pod> <port>  # then curl the probe path yourself
```

## Fix

Hit the probe endpoint by hand. Wrong port/path is the classic; too-aggressive
`initialDelaySeconds`/`failureThreshold` is second. Verify the app actually
listens where the probe looks.

## When to ignore

During known-slow startups (migrations, cache warmup) where the probe is
expected to fail for a while:

```toml
[[ignore]]
finding = "pod_probe_failing"
object = "default/importer-*"
reason = "slow warmup, probes pass after ~10m"
```

## What kubot can't see

Application-level readiness (connected to its database? migrations done?) —
probes are the app's own statement about that, and kubot trusts them.
