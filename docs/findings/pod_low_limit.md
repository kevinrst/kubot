# pod_low_limit

A container memory limit below 32Mi — more likely a typo than a budget.

Severity: warning

## What kubot saw

Memory limit under the floor, with the floor in evidence so the judgment is
auditable.

## Why it matters

Almost no real runtime survives under 32Mi. The container will OOM on the
first thing it does, and the failure looks like an app crash rather than a
config error.

## Verify it yourself

```sh
kubectl get pod <pod> -o jsonpath='{.spec.containers[*].resources.limits}'
```

## Fix

Raise it (or remove it and observe actual usage via `kubot resources`), unless
you genuinely run something that fits — a static binary sipping memory.

## When to ignore

Genuinely tiny workloads:

```toml
[[ignore]]
finding = "pod_low_limit"
object = "default/tiny-proxy"
reason = "static binary, uses 9Mi steady-state"
```

## What kubot can't see

Intent — kubot can't tell a typo (`16Mi` meant `16Gi`) from a deliberate
tight limit. The floor is deliberately low to only catch the absurd.
